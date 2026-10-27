package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedFreshDatabase(t *testing.T) {
	database := os.Getenv("SPORTIK_CATALOG_SEED_TEST_DB")
	if database == "" {
		t.Skip("set SPORTIK_CATALOG_SEED_TEST_DB to a new disposable database on 127.0.0.1:55439")
	}
	require.True(t, strings.HasPrefix(database, "sportik_catalog_test_"))
	ctx := context.Background()
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.Config{
		Host: "127.0.0.1", Port: "55439", User: "sportik_test", Database: database, Timeout: 10 * time.Second,
	})
	require.NoError(t, err)
	defer pool.Close()
	var empty bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT to_regnamespace('sportapp') IS NULL`).Scan(&empty))
	require.True(t, empty, "existing data is never cleared")
	files, err := filepath.Glob("../../migrations/*.up.sql")
	require.NoError(t, err)
	for _, path := range files {
		sql, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, string(sql))
		require.NoError(t, err)
	}
	user, err := users_postgres_repository.NewUsersRepository(pool).CreateUser(ctx, "seed-test@example.test", "test-hash")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `CREATE FUNCTION sportapp.test_seed_delay() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN PERFORM pg_sleep(0.005); RETURN NEW; END $$;
		CREATE TRIGGER test_seed_delay BEFORE INSERT ON sportapp.exercises
		FOR EACH ROW EXECUTE FUNCTION sportapp.test_seed_delay()`)
	require.NoError(t, err)
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for range 4 {
		wg.Go(func() { errs <- run(ctx, pool, user.ID, "Тестовый автор") })
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	var exercises int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&exercises))
	assert.Equal(t, 67, exercises)
	var exerciseID int
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM sportapp.exercises WHERE exercises_name = 'Жим ногами лёжа'`).Scan(&exerciseID))
	require.NoError(t, run(ctx, pool, user.ID, "Тестовый автор"))
	_, err = pool.Exec(ctx, `DELETE FROM sportapp.exercise_equipment
		WHERE exercise_id IN (SELECT id FROM sportapp.exercises WHERE exercises_name = 'Молотки');
		DELETE FROM sportapp.exercises WHERE exercises_name = 'Молотки'`)
	require.NoError(t, err)
	require.NoError(t, run(ctx, pool, user.ID, "Тестовый автор"))
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&exercises))
	assert.Equal(t, 67, exercises)
	var sameID, programs int
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM sportapp.exercises WHERE exercises_name = 'Жим ногами лёжа'`).Scan(&sameID))
	assert.Equal(t, exerciseID, sameID)
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.programs`).Scan(&programs))
	assert.Equal(t, 2, programs)
}
