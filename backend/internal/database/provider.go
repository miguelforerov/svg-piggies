package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/bob"
	bobpgx "github.com/stephenafamo/bob/drivers/pgx"
)

// Querier is the PostgreSQL surface used by repositories. Both pgx.Conn and a
// connection acquired from pgxpool implement it.
type Querier interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Connection struct {
	Querier
	bobTransactor bob.Transactor[bobpgx.Tx]
	release       func(context.Context) error
}

func NewConnection(
	querier Querier,
	bobTransactor bob.Transactor[bobpgx.Tx],
	release func(context.Context) error,
) (*Connection, error) {
	if querier == nil {
		return nil, errors.New("database querier is required")
	}
	if bobTransactor == nil {
		return nil, errors.New("Bob database transactor is required")
	}
	if release == nil {
		return nil, errors.New("database release function is required")
	}
	return &Connection{
		Querier:       querier,
		bobTransactor: bobTransactor,
		release:       release,
	}, nil
}

func (c *Connection) BobTransactor() bob.Transactor[bobpgx.Tx] {
	return c.bobTransactor
}

func (c *Connection) Release(ctx context.Context) error {
	return c.release(ctx)
}

// RunInBobTransaction commits when fn succeeds and rolls back when it fails.
// The callback receives Bob's driver-independent transaction interface.
func RunInBobTransaction[Tx bob.Transaction](
	ctx context.Context,
	transactor bob.Transactor[Tx],
	fn func(context.Context, bob.Transaction) error,
) error {
	transaction, err := transactor.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = transaction.Rollback(context.WithoutCancel(ctx))
	}()

	if err := fn(ctx, transaction); err != nil {
		return err
	}
	return transaction.Commit(ctx)
}

// Provider acquires a connection for one repository operation. The native API
// acquires it from pgxpool; the Worker opens it through Hyperdrive.
type Provider interface {
	Acquire(ctx context.Context) (*Connection, error)
}
