//go:build !js

package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolProvider struct {
	pool *pgxpool.Pool
}

func NewPoolProvider(ctx context.Context, connectionString string) (*PoolProvider, error) {
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL connection string: %w", err)
	}

	// Avoid relying on server-side prepared statements. This also keeps queries
	// compatible when the same repository is used through Hyperdrive.
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	poolConfig.MaxConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	return &PoolProvider{pool: pool}, nil
}

func (p *PoolProvider) Acquire(ctx context.Context) (*Connection, error) {
	connection, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire PostgreSQL connection: %w", err)
	}

	return NewConnection(connection, func(context.Context) error {
		connection.Release()
		return nil
	})
}

func (p *PoolProvider) Close() {
	p.pool.Close()
}
