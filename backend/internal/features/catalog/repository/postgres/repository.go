package catalog_postgres_repository

import core_postgres_pool "sport_app/internal/core/repository/postgres/pool"

type Repository struct {
	pool core_postgres_pool.Pool
}

func NewRepository(pool core_postgres_pool.Pool) *Repository {
	return &Repository{pool: pool}
}
