package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

func main() {
	trainerID := flag.Int("trainer-id", 0, "Registered author user id")
	displayName := flag.String("display-name", "", "Author display name")
	flag.Parse()
	if *trainerID <= 0 || *displayName == "" {
		fmt.Fprintln(os.Stderr, "Specify -trainer-id and -display-name")
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	config, err := core_pgx_pool.NewConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid PostgreSQL configuration")
		os.Exit(1)
	}
	pool, err := core_pgx_pool.NewPool(ctx, config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot connect to PostgreSQL")
		os.Exit(1)
	}
	if err := run(ctx, pool, *trainerID, *displayName); err != nil {
		pool.Close()
		fmt.Fprintln(os.Stderr, "Catalog seed failed; check migrations, author and exercise reference data")
		os.Exit(1)
	}
	pool.Close()
	fmt.Println("Catalog seed completed")
}

func run(ctx context.Context, pool *core_pgx_pool.Pool, trainerID int, displayName string) error {
	if err := users_postgres_repository.NewUsersRepository(pool).EnsureExercisesSeeded(ctx); err != nil {
		return err
	}
	repo := catalog_postgres_repository.NewRepository(pool)
	if err := repo.EnsureReferenceData(ctx); err != nil {
		return err
	}
	return repo.SeedCatalog(ctx, trainerID, displayName)
}
