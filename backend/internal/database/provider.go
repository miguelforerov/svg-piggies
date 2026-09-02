package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	release func(context.Context) error
}

func NewConnection(querier Querier, release func(context.Context) error) (*Connection, error) {
	if querier == nil {
		return nil, errors.New("database querier is required")
	}
	if release == nil {
		return nil, errors.New("database release function is required")
	}
	return &Connection{Querier: querier, release: release}, nil
}

func (c *Connection) Release(ctx context.Context) error {
	return c.release(ctx)
}

// Provider acquires a connection for one repository operation. The native API
// acquires it from pgxpool; the Worker opens it through Hyperdrive.
type Provider interface {
	Acquire(ctx context.Context) (*Connection, error)
}
