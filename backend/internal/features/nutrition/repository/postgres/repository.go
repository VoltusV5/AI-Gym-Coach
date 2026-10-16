package postgres

import (
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
)

// Repository implements the data access layer for nutrition features.
type Repository struct {
	pool core_postgres_pool.Pool
}

// NewRepository creates a new nutrition postgres repository instance.
func NewRepository(pool core_postgres_pool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}
