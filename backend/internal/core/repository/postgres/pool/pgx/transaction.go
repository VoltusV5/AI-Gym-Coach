package core_pgx_pool

import (
	"context"

	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

	"github.com/jackc/pgx/v5"
)

type transaction struct {
	pgx.Tx
}

func (p *Pool) Begin(ctx context.Context) (core_postgres_pool.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return transaction{tx}, nil
}

func (t transaction) Query(ctx context.Context, sql string, args ...any) (core_postgres_pool.Rows, error) {
	rows, err := t.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgxRows{rows}, nil
}

func (t transaction) QueryRow(ctx context.Context, sql string, args ...any) core_postgres_pool.Row {
	return pgxRow{t.Tx.QueryRow(ctx, sql, args...)}
}

func (t transaction) Exec(ctx context.Context, sql string, args ...any) (core_postgres_pool.CommandTag, error) {
	tag, err := t.Tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgxCommandTag{tag}, nil
}
