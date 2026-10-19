package postgres

import (
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

// Repository задаёт доступ к данным питания.
type Repository struct {
	pool core_postgres_pool.Pool
}

// NewRepository создаёт PostgreSQL-репозиторий питания с переданным пулом подключений.
func NewRepository(pool core_postgres_pool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}
