package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "sport_app/internal/core/errors"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	nutrition_repo "sport_app/internal/features/nutrition/repository/postgres"
	nutrition_service "sport_app/internal/features/nutrition/service"
	nutrition_http "sport_app/internal/features/nutrition/transport/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockNutritionService struct {
	ListEntriesFunc     func(ctx context.Context, userID int, dayRaw string, limit, offset int) ([]nutrition_repo.NutritionEntry, int, string, error)
	CreateEntryFunc     func(ctx context.Context, userID int, inp nutrition_service.CreateEntryInput) (nutrition_repo.NutritionEntry, error)
	UpdateEntryFunc     func(ctx context.Context, userID int, entryID int, inp nutrition_service.UpdateEntryInput) error
	DeleteEntryFunc     func(ctx context.Context, userID int, entryID int) error
	ListFavoritesFunc   func(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.NutritionFavorite, int, error)
	CreateFavoriteFunc  func(ctx context.Context, userID int, title string, p, f, c float64, unitType string) (int, error)
	DeleteFavoriteFunc  func(ctx context.Context, userID int, favID int) error
	GetGoalFunc         func(ctx context.Context, userID int) (nutrition_repo.NutritionGoal, error)
	RecalculateGoalFunc func(ctx context.Context, userID int, target string, deltaKg *float64) (nutrition_repo.NutritionGoal, error)
	GetDashboardFunc    func(ctx context.Context, userID int, dayRaw string) (nutrition_service.DashboardResult, error)
	GetStatsFunc        func(ctx context.Context, userID int) (nutrition_service.StatsResult, error)
	GetAnalyticsFunc    func(ctx context.Context, userID int, days int) (nutrition_service.AnalyticsResult, error)
	SearchDishesFunc    func(ctx context.Context, query string, limit, offset int) ([]nutrition_repo.CatalogDish, int, error)
	CreateDishFunc      func(ctx context.Context, userID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error)
	ListMyDishesFunc    func(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.CatalogDish, int, error)
	UpdateMyDishFunc    func(ctx context.Context, userID int, dishID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error)
	DeleteMyDishFunc    func(ctx context.Context, userID int, dishID int) error
	UpsertWaterFunc     func(ctx context.Context, userID int, dayRaw string, amountML int) error
	UpsertWeightFunc    func(ctx context.Context, userID int, dayRaw string, weightKg float64) error
}

func (m *mockNutritionService) ListEntries(ctx context.Context, userID int, dayRaw string, limit, offset int) ([]nutrition_repo.NutritionEntry, int, string, error) {
	if m.ListEntriesFunc != nil {
		return m.ListEntriesFunc(ctx, userID, dayRaw, limit, offset)
	}
	return nil, 0, "2026-10-13", nil
}
func (m *mockNutritionService) CreateEntry(ctx context.Context, userID int, inp nutrition_service.CreateEntryInput) (nutrition_repo.NutritionEntry, error) {
	if m.CreateEntryFunc != nil {
		return m.CreateEntryFunc(ctx, userID, inp)
	}
	return nutrition_repo.NutritionEntry{}, nil
}
func (m *mockNutritionService) UpdateEntry(ctx context.Context, userID int, entryID int, inp nutrition_service.UpdateEntryInput) error {
	if m.UpdateEntryFunc != nil {
		return m.UpdateEntryFunc(ctx, userID, entryID, inp)
	}
	return nil
}
func (m *mockNutritionService) DeleteEntry(ctx context.Context, userID int, entryID int) error {
	if m.DeleteEntryFunc != nil {
		return m.DeleteEntryFunc(ctx, userID, entryID)
	}
	return nil
}
func (m *mockNutritionService) ListFavorites(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.NutritionFavorite, int, error) {
	if m.ListFavoritesFunc != nil {
		return m.ListFavoritesFunc(ctx, userID, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockNutritionService) CreateFavorite(ctx context.Context, userID int, title string, p, f, c float64, unitType string) (int, error) {
	if m.CreateFavoriteFunc != nil {
		return m.CreateFavoriteFunc(ctx, userID, title, p, f, c, unitType)
	}
	return 1, nil
}
func (m *mockNutritionService) DeleteFavorite(ctx context.Context, userID int, favID int) error {
	if m.DeleteFavoriteFunc != nil {
		return m.DeleteFavoriteFunc(ctx, userID, favID)
	}
	return nil
}
func (m *mockNutritionService) GetGoal(ctx context.Context, userID int) (nutrition_repo.NutritionGoal, error) {
	if m.GetGoalFunc != nil {
		return m.GetGoalFunc(ctx, userID)
	}
	return nutrition_repo.NutritionGoal{}, nil
}
func (m *mockNutritionService) RecalculateGoal(ctx context.Context, userID int, target string, deltaKg *float64) (nutrition_repo.NutritionGoal, error) {
	if m.RecalculateGoalFunc != nil {
		return m.RecalculateGoalFunc(ctx, userID, target, deltaKg)
	}
	return nutrition_repo.NutritionGoal{}, nil
}
func (m *mockNutritionService) GetDashboard(ctx context.Context, userID int, dayRaw string) (nutrition_service.DashboardResult, error) {
	if m.GetDashboardFunc != nil {
		return m.GetDashboardFunc(ctx, userID, dayRaw)
	}
	return nutrition_service.DashboardResult{}, nil
}
func (m *mockNutritionService) GetStats(ctx context.Context, userID int) (nutrition_service.StatsResult, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx, userID)
	}
	return nutrition_service.StatsResult{}, nil
}
func (m *mockNutritionService) GetAnalytics(ctx context.Context, userID int, days int) (nutrition_service.AnalyticsResult, error) {
	if m.GetAnalyticsFunc != nil {
		return m.GetAnalyticsFunc(ctx, userID, days)
	}
	return nutrition_service.AnalyticsResult{}, nil
}
func (m *mockNutritionService) SearchDishes(ctx context.Context, query string, limit, offset int) ([]nutrition_repo.CatalogDish, int, error) {
	if m.SearchDishesFunc != nil {
		return m.SearchDishesFunc(ctx, query, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockNutritionService) CreateDish(ctx context.Context, userID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error) {
	if m.CreateDishFunc != nil {
		return m.CreateDishFunc(ctx, userID, dish)
	}
	return dish, nil
}
func (m *mockNutritionService) ListMyDishes(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.CatalogDish, int, error) {
	if m.ListMyDishesFunc != nil {
		return m.ListMyDishesFunc(ctx, userID, limit, offset)
	}
	return nil, 0, nil
}
func (m *mockNutritionService) UpdateMyDish(ctx context.Context, userID int, dishID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error) {
	if m.UpdateMyDishFunc != nil {
		return m.UpdateMyDishFunc(ctx, userID, dishID, dish)
	}
	return dish, nil
}
func (m *mockNutritionService) DeleteMyDish(ctx context.Context, userID int, dishID int) error {
	if m.DeleteMyDishFunc != nil {
		return m.DeleteMyDishFunc(ctx, userID, dishID)
	}
	return nil
}
func (m *mockNutritionService) UpsertWater(ctx context.Context, userID int, dayRaw string, amountML int) error {
	if m.UpsertWaterFunc != nil {
		return m.UpsertWaterFunc(ctx, userID, dayRaw, amountML)
	}
	return nil
}
func (m *mockNutritionService) UpsertWeight(ctx context.Context, userID int, dayRaw string, weightKg float64) error {
	if m.UpsertWeightFunc != nil {
		return m.UpsertWeightFunc(ctx, userID, dayRaw, weightKg)
	}
	return nil
}

func contextWithUser(userID string) context.Context {
	return context.WithValue(context.Background(), core_http_middleware.UserIDKey, userID)
}

func TestNutritionHTTP_ListEntries(t *testing.T) {
	t.Run("Authorized request returns 200 with items and pagination", func(t *testing.T) {
		mockSvc := &mockNutritionService{
			ListEntriesFunc: func(ctx context.Context, userID int, dayRaw string, limit, offset int) ([]nutrition_repo.NutritionEntry, int, string, error) {
				assert.Equal(t, 10, userID)
				assert.Equal(t, 5, limit)
				assert.Equal(t, 10, offset)
				return []nutrition_repo.NutritionEntry{
					{ID: 1, Title: "Oatmeal", Calories: 350},
				}, 1, "2026-10-13", nil
			},
		}

		handler := nutrition_http.NewNutritionHTTPHandler(mockSvc)
		req := httptest.NewRequest(http.MethodGet, "/nutrition/entries?limit=5&offset=10", nil)
		req = req.WithContext(contextWithUser("10"))
		rec := httptest.NewRecorder()

		handler.ListEntries(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]any
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, float64(1), resp["total"])
		assert.Equal(t, float64(5), resp["limit"])
		assert.Equal(t, float64(10), resp["offset"])
	})

	t.Run("Unauthorized request returns 401", func(t *testing.T) {
		mockSvc := &mockNutritionService{}
		handler := nutrition_http.NewNutritionHTTPHandler(mockSvc)
		req := httptest.NewRequest(http.MethodGet, "/nutrition/entries", nil)
		rec := httptest.NewRecorder()

		handler.ListEntries(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestNutritionHTTP_CreateEntry(t *testing.T) {
	t.Run("Valid payload creates entry and returns 201", func(t *testing.T) {
		mockSvc := &mockNutritionService{
			CreateEntryFunc: func(ctx context.Context, userID int, inp nutrition_service.CreateEntryInput) (nutrition_repo.NutritionEntry, error) {
				assert.Equal(t, 10, userID)
				assert.Equal(t, "Chicken Breast", inp.Title)
				return nutrition_repo.NutritionEntry{
					ID:       101,
					Title:    inp.Title,
					Calories: 165,
				}, nil
			},
		}

		handler := nutrition_http.NewNutritionHTTPHandler(mockSvc)
		body := `{"title": "Chicken Breast", "meal_type": "lunch", "protein_g": 31, "fat_g": 3.6, "carbs_g": 0, "calories": 165}`
		req := httptest.NewRequest(http.MethodPost, "/nutrition/entries", bytes.NewBufferString(body))
		req = req.WithContext(contextWithUser("10"))
		rec := httptest.NewRecorder()

		handler.CreateEntry(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp nutrition_repo.NutritionEntry
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, 101, resp.ID)
		assert.Equal(t, "Chicken Breast", resp.Title)
	})

	t.Run("Empty title returns 400 Bad Request", func(t *testing.T) {
		mockSvc := &mockNutritionService{}
		handler := nutrition_http.NewNutritionHTTPHandler(mockSvc)
		body := `{"title": "", "meal_type": "lunch"}`
		req := httptest.NewRequest(http.MethodPost, "/nutrition/entries", bytes.NewBufferString(body))
		req = req.WithContext(contextWithUser("10"))
		rec := httptest.NewRecorder()

		handler.CreateEntry(rec, req)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestNutritionHTTP_DeleteEntry_IDOR(t *testing.T) {
	t.Run("Attempt to delete another user's entry returns 404 (IDOR prevention)", func(t *testing.T) {
		mockSvc := &mockNutritionService{
			DeleteEntryFunc: func(ctx context.Context, userID int, entryID int) error {
				assert.Equal(t, 99, userID)
				assert.Equal(t, 42, entryID)
				// Simulating entity not found or belonging to another user
				return core_errors.ErrNotFound
			},
		}

		handler := nutrition_http.NewNutritionHTTPHandler(mockSvc)
		req := httptest.NewRequest(http.MethodDelete, "/nutrition/entries/42", nil)
		req.SetPathValue("id", "42")
		req = req.WithContext(contextWithUser("99"))
		rec := httptest.NewRecorder()

		handler.DeleteEntry(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestNutritionHTTP_DeleteMyDish_IDOR(t *testing.T) {
	t.Run("Attempt to delete another user's custom dish returns 404", func(t *testing.T) {
		mockSvc := &mockNutritionService{
			DeleteMyDishFunc: func(ctx context.Context, userID int, dishID int) error {
				return core_errors.ErrNotFound
			},
		}

		handler := nutrition_http.NewNutritionHTTPHandler(mockSvc)
		req := httptest.NewRequest(http.MethodDelete, "/nutrition/dishes/55", nil)
		req.SetPathValue("id", "55")
		req = req.WithContext(contextWithUser("12"))
		rec := httptest.NewRecorder()

		handler.DeleteMyDish(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
