package catalog_service

import (
	"context"
	"errors"
	"testing"

	core_errors "sport_app/internal/core/errors"
	catalog_postgres_repository "sport_app/internal/features/catalog/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRepository struct {
	Repository
	filter           catalog_postgres_repository.Filter
	items            []catalog_postgres_repository.ProgramCard
	err              error
	unknownEquipment bool
	calls            int
}

func (r *testRepository) ListPrograms(_ context.Context, filter catalog_postgres_repository.Filter) ([]catalog_postgres_repository.ProgramCard, error) {
	r.calls++
	r.filter = filter
	return r.items, r.err
}

func (r *testRepository) EquipmentExists(context.Context, []string) (bool, error) {
	return !r.unknownEquipment, r.err
}

func TestCatalogPagination(t *testing.T) {
	repo := &testRepository{items: []catalog_postgres_repository.ProgramCard{
		{ID: "18000000-0000-4000-8000-000000000001"}, {ID: "18000000-0000-4000-8000-000000000002"},
	}}
	svc := NewService(repo)
	filter := catalog_postgres_repository.Filter{Limit: 1, Equipment: []string{"dumbbell", "bench"}, EquipmentSet: true}
	page, err := svc.ListPrograms(context.Background(), filter, "")
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.NotNil(t, page.NextCursor)
	assert.Equal(t, []string{"dumbbell", "bench"}, filter.Equipment)
	filter.Equipment = []string{"bench", "dumbbell", "bench"}
	repo.items = repo.items[1:]
	last, err := svc.ListPrograms(context.Background(), filter, *page.NextCursor)
	require.NoError(t, err)
	assert.Equal(t, page.Items[0].ID, repo.filter.After)
	assert.Nil(t, last.NextCursor)
	filter.Goal = "different"
	_, err = svc.ListPrograms(context.Background(), filter, *page.NextCursor)
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
	_, _, err = prepareFilter(catalog_postgres_repository.Filter{Limit: 1, Equipment: []string{"bench", "dumbbell"}, EquipmentSet: true}, *page.NextCursor, "exercises")
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
}

func TestCatalogInvalidFilters(t *testing.T) {
	for _, filter := range []catalog_postgres_repository.Filter{
		{Limit: -1}, {Limit: 101}, {DurationWeeks: -1}, {TrainerID: -1}, {Sort: "arbitrary SQL"},
		{PriceType: "unknown"}, {Equipment: []string{""}}, {Equipment: []string{"missing"}},
	} {
		repo := &testRepository{unknownEquipment: true}
		_, err := NewService(repo).ListPrograms(context.Background(), filter, "")
		assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
		assert.Zero(t, repo.calls)
	}
	for _, raw := range []string{"invalid", "e30", "bnVsbA"} {
		repo := &testRepository{}
		_, err := NewService(repo).ListPrograms(context.Background(), catalog_postgres_repository.Filter{}, raw)
		assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
		assert.Zero(t, repo.calls)
	}
}

func TestCatalogEmptyAndDependencyFailure(t *testing.T) {
	repo := &testRepository{}
	svc := NewService(repo)
	page, err := svc.ListPrograms(context.Background(), catalog_postgres_repository.Filter{}, "")
	require.NoError(t, err)
	assert.NotNil(t, page.Items)
	assert.Nil(t, page.NextCursor)
	assert.Equal(t, 20, repo.filter.Limit)
	repo.err = errors.New("database disconnected")
	_, err = svc.ListPrograms(context.Background(), catalog_postgres_repository.Filter{}, "")
	assert.ErrorIs(t, err, repo.err)
	_, err = svc.GetProgram(context.Background(), "bad-id", 0)
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
	_, err = svc.GrantFreeAccess(context.Background(), "bad-id", 10)
	assert.ErrorIs(t, err, core_errors.ErrInvalidArgument)
	_, err = svc.GrantFreeAccess(context.Background(), "18000000-0000-4000-8000-000000000001", 0)
	assert.ErrorIs(t, err, core_errors.ErrUnauthorized)
}
