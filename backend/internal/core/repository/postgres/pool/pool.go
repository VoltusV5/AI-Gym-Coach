// Package core_postgres_pool описывает пул подключений к БД независимо от конкретного драйвера.
package core_postgres_pool

import (
	"context"
	"time"
)

// Pool описывает выполнение запросов, управление подключениями и таймаутами операций БД.
type Pool interface {
	// Query выполняет SQL-запрос и возвращает итератор строк или ошибку БД.
	Query(ctx context.Context, sql string, args ...any) (Rows, error)

	// QueryRow выполняет SQL-запрос, ожидающий не более одной строки.
	QueryRow(ctx context.Context, sql string, args ...any) Row

	// Exec выполняет SQL-команду и возвращает число затронутых строк или ошибку БД.
	Exec(ctx context.Context, sql string, arguments ...any) (CommandTag, error)

	// Close закрывает все подключения пула.
	Close()

	// OpTimeout возвращает таймаут операции БД.
	OpTimeout() time.Duration
}

// Rows описывает итератор строк результата SQL-запроса.
type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

// Row описывает одну строку результата SQL-запроса.
type Row interface {
	Scan(dest ...any) error
}

// CommandTag содержит сведения о выполненной SQL-команде, включая число затронутых строк.
type CommandTag interface {
	RowsAffected() int64
}

type Tx interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TransactionalPool interface {
	Pool
	Begin(ctx context.Context) (Tx, error)
}
