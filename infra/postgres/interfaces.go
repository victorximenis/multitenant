package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolInterface defines the interface for database connection pool operations
type PoolInterface interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
}

// Ensure pgxpool.Pool implements PoolInterface
var _ PoolInterface = (*pgxpool.Pool)(nil)
