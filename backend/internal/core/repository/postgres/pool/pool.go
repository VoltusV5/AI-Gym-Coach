// Package core_postgres_pool defines connection pool abstractions decoupled from specific database drivers.
package core_postgres_pool

import (
	"context"
	"time"
)

// Pool abstracts database connection pooling operations for queries, executions, and timeouts.
type Pool interface {
	// Query executes an SQL query returning multiple rows.
	Query(ctx context.Context, sql string, args ...any) (Rows, error)

	// QueryRow executes an SQL query expected to return at most one row.
	QueryRow(ctx context.Context, sql string, args ...any) Row

	// Exec executes an SQL statement (INSERT, UPDATE, DELETE) returning the affected rows count.
	Exec(ctx context.Context, sql string, arguments ...any) (CommandTag, error)

	// Close terminates all pool connections.
	Close()

	// OpTimeout returns the default database operation deadline.
	OpTimeout() time.Duration
}

// Rows represents an iterator over database query result sets.
type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

// Row represents a single database row result.
type Row interface {
	Scan(dest ...any) error
}

// CommandTag holds metadata about the executed SQL command (e.g., number of affected rows).
type CommandTag interface {
	RowsAffected() int64
}
