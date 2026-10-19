package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	core_auth "sport_app/internal/core/auth"
	core_errors "sport_app/internal/core/errors"
	core_logger "sport_app/internal/core/logger"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	nutrition_repo "sport_app/internal/features/nutrition/repository/postgres"
	nutrition_service "sport_app/internal/features/nutrition/service"
)

const maxPageSize = 200

// Service задаёт бизнес-операции питания, доступные HTTP-обработчикам.
type Service interface {
	ListEntries(ctx context.Context, userID int, dayRaw string, limit, offset int) ([]nutrition_repo.NutritionEntry, int, string, error)
	CreateEntry(ctx context.Context, userID int, inp nutrition_service.CreateEntryInput) (nutrition_repo.NutritionEntry, error)
	UpdateEntry(ctx context.Context, userID int, entryID int, inp nutrition_service.UpdateEntryInput) error
	DeleteEntry(ctx context.Context, userID int, entryID int) error

	ListFavorites(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.NutritionFavorite, int, error)
	CreateFavorite(ctx context.Context, userID int, title string, p, f, c float64, unitType string) (int, error)
	DeleteFavorite(ctx context.Context, userID int, favID int) error

	GetGoal(ctx context.Context, userID int) (nutrition_repo.NutritionGoal, error)
	RecalculateGoal(ctx context.Context, userID int, target string, deltaKg *float64) (nutrition_repo.NutritionGoal, error)

	GetDashboard(ctx context.Context, userID int, dayRaw string) (nutrition_service.DashboardResult, error)
	GetStats(ctx context.Context, userID int) (nutrition_service.StatsResult, error)
	GetAnalytics(ctx context.Context, userID int, days int) (nutrition_service.AnalyticsResult, error)

	SearchDishes(ctx context.Context, query string, limit, offset int) ([]nutrition_repo.CatalogDish, int, error)
	CreateDish(ctx context.Context, userID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error)
	ListMyDishes(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.CatalogDish, int, error)
	UpdateMyDish(ctx context.Context, userID int, dishID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error)
	DeleteMyDish(ctx context.Context, userID int, dishID int) error

	UpsertWater(ctx context.Context, userID int, dayRaw string, amountML int) error
	UpsertWeight(ctx context.Context, userID int, dayRaw string, weightKg float64) error
}

// NutritionHTTPHandler обрабатывает HTTP-запросы дневника питания.
type NutritionHTTPHandler struct {
	svc Service
}

// NewNutritionHTTPHandler создаёт HTTP-обработчик сервиса питания.
func NewNutritionHTTPHandler(svc Service) *NutritionHTTPHandler {
	return &NutritionHTTPHandler{svc: svc}
}

// RegisterRoutes регистрирует HTTP-маршруты компонента с проверкой JWT на защищённых операциях.
func (h *NutritionHTTPHandler) RegisterRoutes(
	jwt *core_auth.JWT,
	register func(method, path string, handler http.Handler),
) {
	protect := core_http_middleware.Protect(jwt)

	register("GET", "/nutrition/entries", protect(http.HandlerFunc(h.ListEntries)))
	register("POST", "/nutrition/entries", protect(http.HandlerFunc(h.CreateEntry)))
	register("PATCH", "/nutrition/entries/{id}", protect(http.HandlerFunc(h.UpdateEntry)))
	register("DELETE", "/nutrition/entries/{id}", protect(http.HandlerFunc(h.DeleteEntry)))

	register("GET", "/nutrition/favorites", protect(http.HandlerFunc(h.ListFavorites)))
	register("POST", "/nutrition/favorites", protect(http.HandlerFunc(h.CreateFavorite)))
	register("DELETE", "/nutrition/favorites/{id}", protect(http.HandlerFunc(h.DeleteFavorite)))

	register("GET", "/nutrition/goals", protect(http.HandlerFunc(h.GetGoal)))
	register("POST", "/nutrition/goals/recalculate", protect(http.HandlerFunc(h.RecalculateGoal)))
	register("GET", "/nutrition/dashboard", protect(http.HandlerFunc(h.GetDashboard)))
	register("GET", "/nutrition/stats", protect(http.HandlerFunc(h.GetStats)))
	register("GET", "/nutrition/analytics", protect(http.HandlerFunc(h.GetAnalytics)))
	register("GET", "/nutrition/dishes/search", protect(http.HandlerFunc(h.SearchDishes)))
	register("GET", "/nutrition/dishes/mine", protect(http.HandlerFunc(h.ListMyDishes)))
	register("POST", "/nutrition/dishes", protect(http.HandlerFunc(h.CreateDish)))
	register("PATCH", "/nutrition/dishes/{id}", protect(http.HandlerFunc(h.PatchMyDish)))
	register("DELETE", "/nutrition/dishes/{id}", protect(http.HandlerFunc(h.DeleteMyDish)))
	register("POST", "/nutrition/water", protect(http.HandlerFunc(h.UpsertWater)))
	register("POST", "/nutrition/weight", protect(http.HandlerFunc(h.UpsertWeight)))
}

func parseUserID(r *http.Request) (int, error) {
	raw, err := core_http_middleware.UserIDFromContext(r.Context())
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid user id '%s': %w", raw, core_errors.ErrUnauthorized)
	}
	return id, nil
}

func parsePagination(r *http.Request) (limit int, offset int) {
	q := r.URL.Query()
	if lStr := q.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			if l > maxPageSize {
				l = maxPageSize
			}
			limit = l
		}
	}
	if oStr := q.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}
	return limit, offset
}

type entryPayload struct {
	DishID     *int     `json:"dish_id"`
	Grams      *float64 `json:"grams"`
	MealType   string   `json:"meal_type"`
	Title      string   `json:"title"`
	ProteinG   *float64 `json:"protein_g"`
	FatG       *float64 `json:"fat_g"`
	CarbsG     *float64 `json:"carbs_g"`
	Calories   *float64 `json:"calories"`
	ConsumedAt *string  `json:"consumed_at"`
	Day        string   `json:"day"`
}

// ListEntries возвращает записи питания пользователя за день со стабильной сортировкой и необязательной пагинацией.
func (h *NutritionHTTPHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	day := r.URL.Query().Get("day")
	limit, offset := parsePagination(r)

	items, total, dayYMD, err := h.svc.ListEntries(ctx, userID, day, limit, offset)
	if err != nil {
		resp.ErrorResponse(err, "failed to list nutrition entries")
		return
	}

	result := map[string]any{
		"items": items,
		"day":   dayYMD,
		"total": total,
	}
	if limit > 0 {
		result["limit"] = limit
		result["offset"] = offset
	}

	resp.JSONResponse(result, http.StatusOK)
}

// CreateEntry сохраняет запись питания пользователя с пересчётом БЖУ по массе порции и привязкой к блюду. Возвращает запись либо ошибку.
func (h *NutritionHTTPHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	var p entryPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		resp.ErrorResponse(fmt.Errorf("decode entry body: %w", core_errors.ErrInvalidArgument), "invalid payload")
		return
	}
	if p.ProteinG == nil || p.FatG == nil || p.CarbsG == nil {
		resp.ErrorResponse(fmt.Errorf("protein_g, fat_g, carbs_g are required: %w", core_errors.ErrInvalidArgument), "missing macro fields")
		return
	}

	entry, err := h.svc.CreateEntry(ctx, userID, nutrition_service.CreateEntryInput{
		DishID:     p.DishID,
		Grams:      p.Grams,
		MealType:   p.MealType,
		Title:      p.Title,
		ProteinG:   *p.ProteinG,
		FatG:       *p.FatG,
		CarbsG:     *p.CarbsG,
		Calories:   p.Calories,
		ConsumedAt: p.ConsumedAt,
		Day:        p.Day,
	})
	if err != nil {
		resp.ErrorResponse(err, "failed to create nutrition entry")
		return
	}

	resp.JSONResponse(entry, http.StatusOK)
}

// UpdateEntry изменяет принадлежащую пользователю запись питания. Возвращает запись либо ошибку.
func (h *NutritionHTTPHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	entryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		resp.ErrorResponse(fmt.Errorf("invalid entry id: %w", core_errors.ErrInvalidArgument), "invalid id")
		return
	}

	var p entryPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		resp.ErrorResponse(fmt.Errorf("decode entry body: %w", core_errors.ErrInvalidArgument), "invalid payload")
		return
	}
	if p.ProteinG == nil || p.FatG == nil || p.CarbsG == nil {
		resp.ErrorResponse(fmt.Errorf("protein_g, fat_g, carbs_g are required: %w", core_errors.ErrInvalidArgument), "missing macro fields")
		return
	}

	err = h.svc.UpdateEntry(ctx, userID, entryID, nutrition_service.UpdateEntryInput{
		DishID:     p.DishID,
		Grams:      p.Grams,
		MealType:   p.MealType,
		Title:      p.Title,
		ProteinG:   *p.ProteinG,
		FatG:       *p.FatG,
		CarbsG:     *p.CarbsG,
		Calories:   p.Calories,
		ConsumedAt: p.ConsumedAt,
		Day:        p.Day,
	})
	if err != nil {
		resp.ErrorResponse(err, "failed to update nutrition entry")
		return
	}

	resp.JSONResponse(map[string]any{"ok": true}, http.StatusOK)
}

// DeleteEntry удаляет запись питания только её владельца; при отсутствии записи возвращает ErrNotFound.
func (h *NutritionHTTPHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	entryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		resp.ErrorResponse(fmt.Errorf("invalid entry id: %w", core_errors.ErrInvalidArgument), "invalid id")
		return
	}

	if err := h.svc.DeleteEntry(ctx, userID, entryID); err != nil {
		resp.ErrorResponse(err, "failed to delete nutrition entry")
		return
	}

	resp.JSONResponse(map[string]any{"ok": true}, http.StatusOK)
}

// ListFavorites возвращает избранные продукты пользователя.
func (h *NutritionHTTPHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	limit, offset := parsePagination(r)
	items, total, err := h.svc.ListFavorites(ctx, userID, limit, offset)
	if err != nil {
		resp.ErrorResponse(err, "failed to list favorites")
		return
	}

	result := map[string]any{"items": items, "total": total}
	if limit > 0 {
		result["limit"] = limit
		result["offset"] = offset
	}

	resp.JSONResponse(result, http.StatusOK)
}

// CreateFavorite сохраняет продукт в избранном пользователя.
func (h *NutritionHTTPHandler) CreateFavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	var payload struct {
		Title    string   `json:"title"`
		ProteinG *float64 `json:"protein_g"`
		FatG     *float64 `json:"fat_g"`
		CarbsG   *float64 `json:"carbs_g"`
		UnitType string   `json:"unit_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		resp.ErrorResponse(fmt.Errorf("decode favorite body: %w", core_errors.ErrInvalidArgument), "invalid payload")
		return
	}
	if payload.ProteinG == nil || payload.FatG == nil || payload.CarbsG == nil {
		resp.ErrorResponse(fmt.Errorf("protein_g, fat_g, carbs_g are required: %w", core_errors.ErrInvalidArgument), "missing macro fields")
		return
	}

	id, err := h.svc.CreateFavorite(ctx, userID, payload.Title, *payload.ProteinG, *payload.FatG, *payload.CarbsG, payload.UnitType)
	if err != nil {
		resp.ErrorResponse(err, "failed to create favorite")
		return
	}

	resp.JSONResponse(map[string]any{"id": id, "ok": true}, http.StatusOK)
}

// DeleteFavorite удаляет продукт из избранного только его владельца.
func (h *NutritionHTTPHandler) DeleteFavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	favID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		resp.ErrorResponse(fmt.Errorf("invalid favorite id: %w", core_errors.ErrInvalidArgument), "invalid id")
		return
	}

	if err := h.svc.DeleteFavorite(ctx, userID, favID); err != nil {
		resp.ErrorResponse(err, "failed to delete favorite")
		return
	}

	resp.JSONResponse(map[string]any{"ok": true}, http.StatusOK)
}

// GetGoal возвращает цель питания пользователя; сервис рассчитывает и сохраняет её по профилю при отсутствии.
func (h *NutritionHTTPHandler) GetGoal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	goal, err := h.svc.GetGoal(ctx, userID)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			resp.NoContentResponse()
			return
		}
		resp.ErrorResponse(err, "failed to load nutrition goal")
		return
	}

	resp.JSONResponse(goal, http.StatusOK)
}

// RecalculateGoal пересчитывает цели питания по профилю и необязательному изменению массы.
func (h *NutritionHTTPHandler) RecalculateGoal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	var req struct {
		Target        string   `json:"target"`
		TargetDeltaKg *float64 `json:"target_delta_kg"`
	}
	bodyBytes, _ := io.ReadAll(r.Body)
	if len(strings.TrimSpace(string(bodyBytes))) > 0 {
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			resp.ErrorResponse(fmt.Errorf("decode recalculate goal: %w", core_errors.ErrInvalidArgument), "invalid payload")
			return
		}
	}

	goal, err := h.svc.RecalculateGoal(ctx, userID, req.Target, req.TargetDeltaKg)
	if err != nil {
		if errors.Is(err, nutrition_service.ErrProfileIncomplete) {
			resp.ErrorResponse(err, "incomplete profile")
			return
		}
		resp.ErrorResponse(err, "failed to save nutrition goal")
		return
	}

	resp.JSONResponse(goal, http.StatusOK)
}

// GetStats возвращает суммарные и средние значения БЖУ и калорий за день, неделю и месяц.
func (h *NutritionHTTPHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	stats, err := h.svc.GetStats(ctx, userID)
	if err != nil {
		resp.ErrorResponse(err, "failed to build nutrition stats")
		return
	}

	resp.JSONResponse(stats, http.StatusOK)
}

// SearchDishes ищет блюда по подстроке названия со стабильной сортировкой.
func (h *NutritionHTTPHandler) SearchDishes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	query := r.URL.Query().Get("q")
	limit, offset := parsePagination(r)
	if limit == 0 {
		limit = nutrition_service.DefaultDishSearchLimit
	}

	items, total, err := h.svc.SearchDishes(ctx, query, limit, offset)
	if err != nil {
		resp.ErrorResponse(err, "failed to search dishes")
		return
	}

	resp.JSONResponse(map[string]any{"items": items, "total": total}, http.StatusOK)
}

// CreateDish создаёт или обновляет пользовательское блюдо в каталоге.
func (h *NutritionHTTPHandler) CreateDish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	var payload struct {
		Title     string   `json:"title"`
		ProteinG  *float64 `json:"protein_g"`
		FatG      *float64 `json:"fat_g"`
		CarbsG    *float64 `json:"carbs_g"`
		Calories  *float64 `json:"calories"`
		BaseGrams *float64 `json:"base_grams"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		resp.ErrorResponse(fmt.Errorf("decode create dish body: %w", core_errors.ErrInvalidArgument), "invalid payload")
		return
	}

	title := strings.TrimSpace(payload.Title)
	if title == "" {
		resp.ErrorResponse(fmt.Errorf("title is required: %w", core_errors.ErrInvalidArgument), "title is required")
		return
	}

	dish := nutrition_repo.CatalogDish{
		Title: title,
	}
	if payload.ProteinG != nil {
		dish.ProteinG = *payload.ProteinG
	}
	if payload.FatG != nil {
		dish.FatG = *payload.FatG
	}
	if payload.CarbsG != nil {
		dish.CarbsG = *payload.CarbsG
	}
	if payload.Calories != nil {
		dish.Calories = *payload.Calories
	}
	if payload.BaseGrams != nil {
		dish.BaseGrams = *payload.BaseGrams
	}

	created, err := h.svc.CreateDish(ctx, userID, dish)
	if err != nil {
		resp.ErrorResponse(err, "failed to create dish")
		return
	}

	resp.JSONResponse(created, http.StatusOK)
}

// ListMyDishes возвращает блюда, созданные текущим пользователем.
func (h *NutritionHTTPHandler) ListMyDishes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	limit, offset := parsePagination(r)
	items, total, err := h.svc.ListMyDishes(ctx, userID, limit, offset)
	if err != nil {
		resp.ErrorResponse(err, "failed to list my dishes")
		return
	}

	resp.JSONResponse(map[string]any{"items": items, "total": total}, http.StatusOK)
}

// PatchMyDish обрабатывает изменение блюда только его автором.
func (h *NutritionHTTPHandler) PatchMyDish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	dishID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || dishID <= 0 {
		resp.ErrorResponse(fmt.Errorf("invalid dish id: %w", core_errors.ErrInvalidArgument), "invalid id")
		return
	}

	var payload struct {
		Title     string   `json:"title"`
		ProteinG  *float64 `json:"protein_g"`
		FatG      *float64 `json:"fat_g"`
		CarbsG    *float64 `json:"carbs_g"`
		Calories  *float64 `json:"calories"`
		BaseGrams *float64 `json:"base_grams"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		resp.ErrorResponse(fmt.Errorf("decode dish body: %w", core_errors.ErrInvalidArgument), "invalid payload")
		return
	}

	dish := nutrition_repo.CatalogDish{
		Title: strings.TrimSpace(payload.Title),
	}
	if payload.ProteinG != nil {
		dish.ProteinG = *payload.ProteinG
	}
	if payload.FatG != nil {
		dish.FatG = *payload.FatG
	}
	if payload.CarbsG != nil {
		dish.CarbsG = *payload.CarbsG
	}
	if payload.Calories != nil {
		dish.Calories = *payload.Calories
	}
	if payload.BaseGrams != nil {
		dish.BaseGrams = *payload.BaseGrams
	}

	updated, err := h.svc.UpdateMyDish(ctx, userID, dishID, dish)
	if err != nil {
		resp.ErrorResponse(err, "failed to update dish")
		return
	}

	resp.JSONResponse(updated, http.StatusOK)
}

// DeleteMyDish удаляет блюдо только для его автора.
func (h *NutritionHTTPHandler) DeleteMyDish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	dishID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || dishID <= 0 {
		resp.ErrorResponse(fmt.Errorf("invalid dish id: %w", core_errors.ErrInvalidArgument), "invalid id")
		return
	}

	if err := h.svc.DeleteMyDish(ctx, userID, dishID); err != nil {
		resp.ErrorResponse(err, "failed to delete dish")
		return
	}

	resp.NoContentResponse()
}

// UpsertWater сохраняет объём выпитой воды пользователя за указанную дату.
func (h *NutritionHTTPHandler) UpsertWater(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	var payload struct {
		AmountML int    `json:"amount_ml"`
		Day      string `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		resp.ErrorResponse(fmt.Errorf("decode water body: %w", core_errors.ErrInvalidArgument), "invalid payload")
		return
	}

	if err := h.svc.UpsertWater(ctx, userID, payload.Day, payload.AmountML); err != nil {
		resp.ErrorResponse(err, "failed to save water")
		return
	}

	resp.JSONResponse(map[string]any{"ok": true}, http.StatusOK)
}

// UpsertWeight сохраняет массу тела пользователя за указанную дату.
func (h *NutritionHTTPHandler) UpsertWeight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	var payload struct {
		WeightKg float64 `json:"weight_kg"`
		Day      string  `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		resp.ErrorResponse(fmt.Errorf("decode weight body: %w", core_errors.ErrInvalidArgument), "invalid payload")
		return
	}
	if payload.WeightKg <= 0 {
		resp.ErrorResponse(fmt.Errorf("weight_kg must be > 0: %w", core_errors.ErrInvalidArgument), "invalid weight")
		return
	}

	if err := h.svc.UpsertWeight(ctx, userID, payload.Day, payload.WeightKg); err != nil {
		resp.ErrorResponse(err, "failed to save weight")
		return
	}

	resp.JSONResponse(map[string]any{"ok": true}, http.StatusOK)
}

// GetDashboard объединяет суточные показатели питания, цели, воду, вес и непрерывную серию записей.
func (h *NutritionHTTPHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	day := r.URL.Query().Get("day")
	dashboard, err := h.svc.GetDashboard(ctx, userID, day)
	if err != nil {
		resp.ErrorResponse(err, "failed to build dashboard")
		return
	}

	resp.JSONResponse(dashboard, http.StatusOK)
}

// GetAnalytics возвращает историю питания, веса и воды за выбранный период.
func (h *NutritionHTTPHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_responce.NewHTTPResponce(log, w)

	userID, err := parseUserID(r)
	if err != nil {
		resp.ErrorResponse(err, "unauthorized")
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	analytics, err := h.svc.GetAnalytics(ctx, userID, days)
	if err != nil {
		resp.ErrorResponse(err, "failed to build analytics")
		return
	}

	resp.JSONResponse(analytics, http.StatusOK)
}
