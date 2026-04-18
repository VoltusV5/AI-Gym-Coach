package nutrition

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	middleware "sport_app/internal/core/transport/http/middleware"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Календарный день дневника питания — Europe/Moscow (как у пользователей приложения).
var nutritionTZ *time.Location

func init() {
	var err error
	nutritionTZ, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		nutritionTZ = time.UTC
	}
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	router.Handle("/api/v1/nutrition/entries", middleware.Protect()(http.HandlerFunc(s.listEntries))).Methods("GET")
	router.Handle("/api/v1/nutrition/entries", middleware.Protect()(http.HandlerFunc(s.createEntry))).Methods("POST")
	router.Handle("/api/v1/nutrition/entries/{id:[0-9]+}", middleware.Protect()(http.HandlerFunc(s.updateEntry))).Methods("PATCH")
	router.Handle("/api/v1/nutrition/entries/{id:[0-9]+}", middleware.Protect()(http.HandlerFunc(s.deleteEntry))).Methods("DELETE")

	router.Handle("/api/v1/nutrition/favorites", middleware.Protect()(http.HandlerFunc(s.listFavorites))).Methods("GET")
	router.Handle("/api/v1/nutrition/favorites", middleware.Protect()(http.HandlerFunc(s.createFavorite))).Methods("POST")
	router.Handle("/api/v1/nutrition/favorites/{id:[0-9]+}", middleware.Protect()(http.HandlerFunc(s.deleteFavorite))).Methods("DELETE")

	router.Handle("/api/v1/nutrition/goals", middleware.Protect()(http.HandlerFunc(s.getGoal))).Methods("GET")
	router.Handle("/api/v1/nutrition/goals/recalculate", middleware.Protect()(http.HandlerFunc(s.recalculateGoal))).Methods("POST")
	router.Handle("/api/v1/nutrition/dashboard", middleware.Protect()(http.HandlerFunc(s.getDashboard))).Methods("GET")
	router.Handle("/api/v1/nutrition/stats", middleware.Protect()(http.HandlerFunc(s.getStats))).Methods("GET")
	router.Handle("/api/v1/nutrition/analytics", middleware.Protect()(http.HandlerFunc(s.getReports))).Methods("GET")
	router.Handle("/api/v1/nutrition/dishes/search", middleware.Protect()(http.HandlerFunc(s.searchDishes))).Methods("GET")
	router.Handle("/api/v1/nutrition/dishes/mine", middleware.Protect()(http.HandlerFunc(s.listMyDishes))).Methods("GET")
	router.Handle("/api/v1/nutrition/dishes", middleware.Protect()(http.HandlerFunc(s.createDish))).Methods("POST")
	router.Handle("/api/v1/nutrition/dishes/{id:[0-9]+}", middleware.Protect()(http.HandlerFunc(s.patchMyDish))).Methods("PATCH")
	router.Handle("/api/v1/nutrition/dishes/{id:[0-9]+}", middleware.Protect()(http.HandlerFunc(s.deleteMyDish))).Methods("DELETE")
	router.Handle("/api/v1/nutrition/water", middleware.Protect()(http.HandlerFunc(s.upsertWater))).Methods("POST")
	router.Handle("/api/v1/nutrition/weight", middleware.Protect()(http.HandlerFunc(s.upsertWeight))).Methods("POST")
}

type nutritionEntry struct {
	ID         int       `json:"id"`
	DishID     *int      `json:"dish_id,omitempty"`
	Grams      float64   `json:"grams"`
	MealType   string    `json:"meal_type"`
	Title      string    `json:"title"`
	ProteinG   float64   `json:"protein_g"`
	FatG       float64   `json:"fat_g"`
	CarbsG     float64   `json:"carbs_g"`
	Calories   float64   `json:"calories"`
	ConsumedAt time.Time `json:"consumed_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type nutritionFavorite struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	ProteinG float64 `json:"protein_g"`
	FatG     float64 `json:"fat_g"`
	CarbsG   float64 `json:"carbs_g"`
	UnitType string  `json:"unit_type"`
}

type nutritionGoal struct {
	ProteinG  float64   `json:"protein_g"`
	FatG      float64   `json:"fat_g"`
	CarbsG    float64   `json:"carbs_g"`
	Calories  float64   `json:"calories"`
	UpdatedAt time.Time `json:"updated_at"`
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
	// Календарный день дневника YYYY-MM-DD (Europe/Moscow, как в GET …/entries?day=).
	// Если задан — consumed_at ставится на полдень этого дня, чтобы запись попала в те же сутки, что и сводка.
	Day string `json:"day"`
}

type recalcPayload struct {
	Target        string   `json:"target"`
	TargetDeltaKg *float64 `json:"target_delta_kg"`
}

// ErrProfileIncomplete зарезервировано для редких случаев (раньше — пустой профиль).
var ErrProfileIncomplete = errors.New("profile incomplete")

func userIDFromRequest(r *http.Request) (int, error) {
	raw, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || raw == "" {
		return 0, errors.New("user id missing")
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func mealTypeOrDefault(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "breakfast":
		return "breakfast"
	case "lunch":
		return "lunch"
	case "dinner":
		return "dinner"
	case "snack":
		return "snack"
	default:
		return "snack"
	}
}

// dayBoundsFromQuery задаёт интервал [startUTC, endUTC) для суток по календарю в nutritionTZ и строку даты YYYY-MM-DD.
func dayBoundsFromQuery(r *http.Request) (startUTC, endUTC time.Time, dayYMD string) {
	dayRaw := strings.TrimSpace(r.URL.Query().Get("day"))
	now := time.Now().In(nutritionTZ)
	var dayStart time.Time
	if dayRaw == "" {
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, nutritionTZ)
	} else if d, err := time.ParseInLocation("2006-01-02", dayRaw, nutritionTZ); err == nil {
		dayStart = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, nutritionTZ)
	} else {
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, nutritionTZ)
	}
	end := dayStart.Add(24 * time.Hour)
	return dayStart.UTC(), end.UTC(), dayStart.Format("2006-01-02")
}

func caloriesFromMacros(p, f, c float64) float64 {
	return p*4 + f*9 + c*4
}

func scaleFrom100(valuePer100, grams float64) float64 {
	if grams <= 0 {
		grams = 100
	}
	return (valuePer100 / 100.0) * grams
}

func (s *Service) listEntries(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	dayStartUTC, dayEndUTC, dayYMD := dayBoundsFromQuery(r)

	rows, err := s.pool.Query(
		r.Context(),
		`SELECT id, dish_id, grams, meal_type, title, protein_g, fat_g, carbs_g, calories, consumed_at, created_at
		 FROM sportapp.nutrition_entries
		 WHERE user_id = $1 AND consumed_at >= $2 AND consumed_at < $3
		 ORDER BY consumed_at DESC, id DESC`,
		userID, dayStartUTC, dayEndUTC,
	)
	if err != nil {
		http.Error(w, "Failed to load entries", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]nutritionEntry, 0, 128)
	for rows.Next() {
		var it nutritionEntry
		if err := rows.Scan(&it.ID, &it.DishID, &it.Grams, &it.MealType, &it.Title, &it.ProteinG, &it.FatG, &it.CarbsG, &it.Calories, &it.ConsumedAt, &it.CreatedAt); err != nil {
			http.Error(w, "Failed to scan entries", http.StatusInternalServerError)
			return
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "day": dayYMD})
}

func parseOptionalTime(v *string) (time.Time, error) {
	if v == nil || strings.TrimSpace(*v) == "" {
		return time.Now().UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*v))
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func consumedAtFromEntryPayload(p *entryPayload) (time.Time, error) {
	if p == nil {
		return time.Now().UTC(), nil
	}
	if strings.TrimSpace(p.Day) != "" {
		d, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(p.Day), nutritionTZ)
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, nutritionTZ).UTC(), nil
	}
	return parseOptionalTime(p.ConsumedAt)
}

func (s *Service) createEntry(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload entryPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if payload.ProteinG == nil || payload.FatG == nil || payload.CarbsG == nil {
		http.Error(w, "protein_g/fat_g/carbs_g are required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Title) == "" {
		payload.Title = "Без названия"
	}
	consumedAt, err := consumedAtFromEntryPayload(&payload)
	if err != nil {
		http.Error(w, "Invalid consumed_at or day", http.StatusBadRequest)
		return
	}

	mealType := mealTypeOrDefault(payload.MealType)
	grams := 100.0
	if payload.Grams != nil && *payload.Grams > 0 {
		grams = *payload.Grams
	}

	var dishID *int
	if payload.DishID != nil && *payload.DishID > 0 {
		var per100P, per100F, per100C, per100K float64
		err := s.pool.QueryRow(
			r.Context(),
			`SELECT protein_g, fat_g, carbs_g, calories FROM sportapp.nutrition_dishes WHERE id = $1`,
			*payload.DishID,
		).Scan(&per100P, &per100F, &per100C, &per100K)
		if err == nil {
			dishID = payload.DishID
			p := scaleFrom100(per100P, grams)
			f := scaleFrom100(per100F, grams)
			c := scaleFrom100(per100C, grams)
			*payload.ProteinG, *payload.FatG, *payload.CarbsG = p, f, c
			if payload.Calories == nil || *payload.Calories <= 0 {
				cal := scaleFrom100(per100K, grams)
				payload.Calories = &cal
			}
		} else {
			// Частый баг фронта: в dish_id передают id избранного/записи дневника. Тогда БЖУ с клиента трактуем как на 100 г.
			pg := scaleFrom100(*payload.ProteinG, grams)
			fg := scaleFrom100(*payload.FatG, grams)
			cg := scaleFrom100(*payload.CarbsG, grams)
			*payload.ProteinG, *payload.FatG, *payload.CarbsG = pg, fg, cg
			var k float64
			if payload.Calories != nil && *payload.Calories > 0 {
				k = scaleFrom100(*payload.Calories, grams)
			} else {
				k = caloriesFromMacros(pg, fg, cg)
			}
			payload.Calories = &k
			dishID = nil
		}
	} else {
		var autoID int
		err = s.pool.QueryRow(
			r.Context(),
			`INSERT INTO sportapp.nutrition_dishes (title, protein_g, fat_g, carbs_g, calories, base_grams, created_by_user_id)
			 VALUES ($1,$2,$3,$4,$5,100,$6)
			 ON CONFLICT (title) DO UPDATE
			 SET protein_g = EXCLUDED.protein_g, fat_g = EXCLUDED.fat_g, carbs_g = EXCLUDED.carbs_g, calories = EXCLUDED.calories, updated_at = NOW()
			 RETURNING id`,
			payload.Title, *payload.ProteinG, *payload.FatG, *payload.CarbsG, caloriesFromMacros(*payload.ProteinG, *payload.FatG, *payload.CarbsG), userID,
		).Scan(&autoID)
		if err == nil {
			dishID = &autoID
			// Клиент присылает БЖУ на 100 г; в nutrition_dishes уже сохранили per-100, в строку дневника — порцию.
			perP := *payload.ProteinG
			perF := *payload.FatG
			perC := *payload.CarbsG
			var perK float64
			if payload.Calories != nil && *payload.Calories > 0 {
				perK = *payload.Calories
			} else {
				perK = caloriesFromMacros(perP, perF, perC)
			}
			*payload.ProteinG = scaleFrom100(perP, grams)
			*payload.FatG = scaleFrom100(perF, grams)
			*payload.CarbsG = scaleFrom100(perC, grams)
			k := scaleFrom100(perK, grams)
			payload.Calories = &k
		}
	}

	calories := caloriesFromMacros(*payload.ProteinG, *payload.FatG, *payload.CarbsG)
	if payload.Calories != nil && *payload.Calories > 0 {
		calories = *payload.Calories
	}

	var id int
	var createdAt time.Time
	err = s.pool.QueryRow(
		r.Context(),
		`INSERT INTO sportapp.nutrition_entries (user_id, dish_id, grams, meal_type, title, protein_g, fat_g, carbs_g, calories, consumed_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 RETURNING id, created_at`,
		userID, dishID, grams, mealType, payload.Title, *payload.ProteinG, *payload.FatG, *payload.CarbsG, calories, consumedAt,
	).Scan(&id, &createdAt)
	if err != nil {
		http.Error(w, "Failed to create entry", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, nutritionEntry{
		ID:         id,
		DishID:     dishID,
		Grams:      grams,
		MealType:   mealType,
		Title:      payload.Title,
		ProteinG:   *payload.ProteinG,
		FatG:       *payload.FatG,
		CarbsG:     *payload.CarbsG,
		Calories:   calories,
		ConsumedAt: consumedAt,
		CreatedAt:  createdAt,
	})
}

func (s *Service) updateEntry(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idRaw := mux.Vars(r)["id"]
	entryID, err := strconv.Atoi(idRaw)
	if err != nil {
		http.Error(w, "Invalid entry id", http.StatusBadRequest)
		return
	}

	var payload entryPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if payload.ProteinG == nil || payload.FatG == nil || payload.CarbsG == nil {
		http.Error(w, "protein_g/fat_g/carbs_g are required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Title) == "" {
		payload.Title = "Без названия"
	}
	consumedAt, err := consumedAtFromEntryPayload(&payload)
	if err != nil {
		http.Error(w, "Invalid consumed_at or day", http.StatusBadRequest)
		return
	}

	mealType := mealTypeOrDefault(payload.MealType)
	grams := 100.0
	if payload.Grams != nil && *payload.Grams > 0 {
		grams = *payload.Grams
	}
	if payload.DishID != nil && *payload.DishID > 0 {
		var per100P, per100F, per100C, per100K float64
		err := s.pool.QueryRow(
			r.Context(),
			`SELECT protein_g, fat_g, carbs_g, calories FROM sportapp.nutrition_dishes WHERE id = $1`,
			*payload.DishID,
		).Scan(&per100P, &per100F, &per100C, &per100K)
		if err == nil {
			p := scaleFrom100(per100P, grams)
			f := scaleFrom100(per100F, grams)
			c := scaleFrom100(per100C, grams)
			*payload.ProteinG, *payload.FatG, *payload.CarbsG = p, f, c
			if payload.Calories == nil || *payload.Calories <= 0 {
				cal := scaleFrom100(per100K, grams)
				payload.Calories = &cal
			}
		} else {
			pg := scaleFrom100(*payload.ProteinG, grams)
			fg := scaleFrom100(*payload.FatG, grams)
			cg := scaleFrom100(*payload.CarbsG, grams)
			*payload.ProteinG, *payload.FatG, *payload.CarbsG = pg, fg, cg
			if payload.Calories != nil && *payload.Calories > 0 {
				k := scaleFrom100(*payload.Calories, grams)
				payload.Calories = &k
			} else {
				k := caloriesFromMacros(pg, fg, cg)
				payload.Calories = &k
			}
		}
	}
	calories := caloriesFromMacros(*payload.ProteinG, *payload.FatG, *payload.CarbsG)
	if payload.Calories != nil && *payload.Calories > 0 {
		calories = *payload.Calories
	}

	tag, err := s.pool.Exec(
		r.Context(),
		`UPDATE sportapp.nutrition_entries
		 SET grams = $1, meal_type = $2, title = $3, protein_g = $4, fat_g = $5, carbs_g = $6, calories = $7, consumed_at = $8, updated_at = NOW()
		 WHERE id = $9 AND user_id = $10`,
		grams, mealType, payload.Title, *payload.ProteinG, *payload.FatG, *payload.CarbsG, calories, consumedAt, entryID, userID,
	)
	if err != nil {
		http.Error(w, "Failed to update entry", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Service) deleteEntry(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idRaw := mux.Vars(r)["id"]
	entryID, err := strconv.Atoi(idRaw)
	if err != nil {
		http.Error(w, "Invalid entry id", http.StatusBadRequest)
		return
	}
	tag, err := s.pool.Exec(r.Context(), `DELETE FROM sportapp.nutrition_entries WHERE id = $1 AND user_id = $2`, entryID, userID)
	if err != nil {
		http.Error(w, "Failed to delete entry", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Service) listFavorites(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := s.pool.Query(
		r.Context(),
		`SELECT id, title, protein_g, fat_g, carbs_g, unit_type
		 FROM sportapp.nutrition_favorites
		 WHERE user_id = $1
		 ORDER BY id DESC`,
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to load favorites", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	items := make([]nutritionFavorite, 0, 128)
	for rows.Next() {
		var it nutritionFavorite
		if err := rows.Scan(&it.ID, &it.Title, &it.ProteinG, &it.FatG, &it.CarbsG, &it.UnitType); err != nil {
			http.Error(w, "Failed to scan favorites", http.StatusInternalServerError)
			return
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Service) createFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if payload.ProteinG == nil || payload.FatG == nil || payload.CarbsG == nil {
		http.Error(w, "protein_g/fat_g/carbs_g are required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Title) == "" {
		payload.Title = "Без названия"
	}
	if payload.UnitType == "" {
		payload.UnitType = "gram"
	}
	var id int
	err = s.pool.QueryRow(
		r.Context(),
		`INSERT INTO sportapp.nutrition_favorites (user_id, title, protein_g, fat_g, carbs_g, unit_type)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		userID, payload.Title, *payload.ProteinG, *payload.FatG, *payload.CarbsG, payload.UnitType,
	).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to create favorite", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "ok": true})
}

func (s *Service) deleteFavorite(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idRaw := mux.Vars(r)["id"]
	favID, err := strconv.Atoi(idRaw)
	if err != nil {
		http.Error(w, "Invalid favorite id", http.StatusBadRequest)
		return
	}
	tag, err := s.pool.Exec(r.Context(), `DELETE FROM sportapp.nutrition_favorites WHERE id = $1 AND user_id = $2`, favID, userID)
	if err != nil {
		http.Error(w, "Failed to delete favorite", http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, "Favorite not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Service) getGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var g nutritionGoal
	err = s.pool.QueryRow(
		r.Context(),
		`SELECT protein_g, fat_g, carbs_g, calories, updated_at
		 FROM sportapp.nutrition_goals WHERE user_id = $1`,
		userID,
	).Scan(&g.ProteinG, &g.FatG, &g.CarbsG, &g.Calories, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			synced, syncErr := s.upsertNutritionGoalFromProfile(r.Context(), userID, "", nil)
			if syncErr != nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			writeJSON(w, http.StatusOK, synced)
			return
		}
		http.Error(w, "Failed to load goal", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

// upsertNutritionGoalFromProfile writes sportapp.nutrition_goals from onboarding profile (Mifflin–St Jeor + macros).
// reqTarget: optional "lose"|"maintain"|"gain" overriding profile.goal wording.
func (s *Service) upsertNutritionGoalFromProfile(ctx context.Context, userID int, reqTarget string, targetDeltaKg *float64) (nutritionGoal, error) {
	type profileRow struct {
		Age           *int
		Gender        *string
		HeightCm      *int
		WeightKg      *int
		ActivityLevel *string
		Goal          *string
	}
	var p profileRow
	err := s.pool.QueryRow(
		ctx,
		`SELECT age, gender, height_cm, weight_kg, activity_level, goal
		 FROM sportapp.profile WHERE user_id = $1`,
		userID,
	).Scan(&p.Age, &p.Gender, &p.HeightCm, &p.WeightKg, &p.ActivityLevel, &p.Goal)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nutritionGoal{}, err
	}

	// Значения по умолчанию, если онбординг не заполнил профиль — иначе GET /goals и дашборд остаются пустыми.
	age := 30
	if p.Age != nil && *p.Age > 0 && *p.Age < 120 {
		age = *p.Age
	}
	heightCm := 175.0
	if p.HeightCm != nil && *p.HeightCm > 80 && *p.HeightCm < 250 {
		heightCm = float64(*p.HeightCm)
	}
	weightKg := 70.0
	if p.WeightKg != nil && *p.WeightKg > 30 && *p.WeightKg < 400 {
		weightKg = float64(*p.WeightKg)
	} else {
		var lw *float64
		if qerr := s.pool.QueryRow(ctx,
			`SELECT weight_kg FROM sportapp.nutrition_weight_logs WHERE user_id = $1 ORDER BY logged_on DESC LIMIT 1`,
			userID,
		).Scan(&lw); qerr == nil && lw != nil && *lw > 30 && *lw < 400 {
			weightKg = *lw
		}
	}

	goal := ""
	if p.Goal != nil {
		goal = *p.Goal
	}
	if strings.TrimSpace(reqTarget) != "" {
		switch reqTarget {
		case "lose":
			goal = "Сбросить вес"
		case "maintain":
			goal = "Поддержание"
		case "gain":
			goal = "Набрать мышцы"
		}
	}
	gender := "Мужчина"
	if p.Gender != nil {
		gender = *p.Gender
	}
	activity := ""
	if p.ActivityLevel != nil {
		activity = *p.ActivityLevel
	}

	newGoal := calculateGoalFromProfile(age, heightCm, weightKg, gender, activity, goal)
	// Только при ненулевой дельте по весу пересчитываем калории и перераспределяем жиры/углеводы (0 в JSON не считаем «изменением»).
	if targetDeltaKg != nil && math.Abs(*targetDeltaKg) > 1e-6 {
		newGoal.Calories += targetDeltaAdjustment(*targetDeltaKg)
		if newGoal.Calories < 1200 {
			newGoal.Calories = 1200
		}
		newGoal.FatG = (newGoal.Calories * 0.28) / 9
		newGoal.CarbsG = (newGoal.Calories - newGoal.ProteinG*4 - newGoal.FatG*9) / 4
		if newGoal.CarbsG < 0 {
			newGoal.CarbsG = 0
		}
	}
	var updatedAt time.Time
	err = s.pool.QueryRow(
		ctx,
		`INSERT INTO sportapp.nutrition_goals (user_id, protein_g, fat_g, carbs_g, calories)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (user_id) DO UPDATE
		 SET protein_g = EXCLUDED.protein_g,
		     fat_g = EXCLUDED.fat_g,
		     carbs_g = EXCLUDED.carbs_g,
		     calories = EXCLUDED.calories,
		     updated_at = NOW()
		 RETURNING updated_at`,
		userID, newGoal.ProteinG, newGoal.FatG, newGoal.CarbsG, newGoal.Calories,
	).Scan(&updatedAt)
	if err != nil {
		return nutritionGoal{}, err
	}
	newGoal.UpdatedAt = updatedAt
	return newGoal, nil
}

func activityMultiplier(level string) float64 {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "сидячий и малоподвижный":
		return 1.2
	case "лёгкая активность (физические нагрузки 1-3 раза в неделю)", "легкая активность (физические нагрузки 1-3 раза в неделю)":
		return 1.375
	case "средняя активность (физические нагрузки 3-5 раза в неделю)":
		return 1.55
	case "высокая активность (физические нагрузки 6-7 раз в неделю)":
		return 1.725
	case "очень высокая активность (постоянно высокая физическая нагрузка)":
		return 1.9
	default:
		return 1.375
	}
}

func targetAdjustment(target string, weightKg float64) float64 {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "lose", "сбросить вес", "сжечь жир":
		return -2 * weightKg
	case "gain", "набрать мышцы", "набрать мышцы и сжечь жир":
		return 2 * weightKg
	default:
		return 0
	}
}

func targetDeltaAdjustment(deltaKg float64) float64 {
	// Soft adjustment for daily calories by desired weight direction.
	return deltaKg * 220
}

// dailyWaterGoalMlFromWeight — суточная норма воды (мл) по массе тела: ~33 мл/кг, с разумными пределами.
func dailyWaterGoalMlFromWeight(weightKg float64) int {
	if weightKg <= 0 {
		return 2000
	}
	ml := int(math.Round(weightKg * 33))
	if ml < 1500 {
		ml = 1500
	}
	if ml > 4500 {
		ml = 4500
	}
	return ml
}

func calculateGoalFromProfile(age int, heightCm, weightKg float64, gender, activity, target string) nutritionGoal {
	base := (10*weightKg + 6.25*heightCm - 5*float64(age))
	if strings.EqualFold(strings.TrimSpace(gender), "Женщина") || strings.EqualFold(strings.TrimSpace(gender), "female") {
		base -= 161
	} else {
		base += 5
	}
	calories := base*activityMultiplier(activity) + targetAdjustment(target, weightKg)
	if calories < 1200 {
		calories = 1200
	}

	var protein float64
	var fat float64
	if strings.Contains(strings.ToLower(target), "сброс") || strings.Contains(strings.ToLower(target), "сжечь") {
		protein = weightKg * 1.2
		fat = (calories * 0.2) / 9
	} else if strings.Contains(strings.ToLower(target), "набрать") {
		protein = weightKg * 1.8
		fat = (calories * 0.25) / 9
	} else {
		protein = weightKg * 1.0
		fat = (calories * 0.3) / 9
	}
	carbs := (calories - protein*4 - fat*9) / 4
	if carbs < 0 {
		carbs = 0
	}
	return nutritionGoal{
		ProteinG: protein,
		FatG:     fat,
		CarbsG:   carbs,
		Calories: calories,
	}
}

func (s *Service) recalculateGoal(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	var req recalcPayload
	if len(bytes.TrimSpace(bodyBytes)) > 0 {
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
	}

	newGoal, err := s.upsertNutritionGoalFromProfile(r.Context(), userID, req.Target, req.TargetDeltaKg)
	if err != nil {
		if errors.Is(err, ErrProfileIncomplete) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"message": "Недостаточно данных для расчёта цели. Заполните профиль или добавьте вес в дневнике.",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Не удалось сохранить цель"})
		return
	}
	writeJSON(w, http.StatusOK, newGoal)
}

func (s *Service) getStats(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	now := time.Now().In(nutritionTZ)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, nutritionTZ)
	weekStart := dayStart.AddDate(0, 0, -7)
	monthStart := dayStart.AddDate(0, 0, -30)

	type sums struct{ protein, fat, carbs, calories float64 }
	sumRange := func(ctx context.Context, from time.Time) (sums, error) {
		var acc sums
		err := s.pool.QueryRow(
			ctx,
			`SELECT COALESCE(SUM(protein_g),0), COALESCE(SUM(fat_g),0), COALESCE(SUM(carbs_g),0), COALESCE(SUM(calories),0)
			 FROM sportapp.nutrition_entries
			 WHERE user_id = $1 AND consumed_at >= $2`,
			userID, from,
		).Scan(&acc.protein, &acc.fat, &acc.carbs, &acc.calories)
		return acc, err
	}

	// «Сегодня» — тот же интервал, что и GET /dashboard?day= (сутки Europe/Moscow).
	dayEnd := dayStart.Add(24 * time.Hour)
	var day sums
	err = s.pool.QueryRow(
		r.Context(),
		`SELECT COALESCE(SUM(protein_g),0), COALESCE(SUM(fat_g),0), COALESCE(SUM(carbs_g),0), COALESCE(SUM(calories),0)
		 FROM sportapp.nutrition_entries
		 WHERE user_id = $1 AND consumed_at >= $2 AND consumed_at < $3`,
		userID, dayStart.UTC(), dayEnd.UTC(),
	).Scan(&day.protein, &day.fat, &day.carbs, &day.calories)
	if err != nil {
		http.Error(w, "Failed to build stats", http.StatusInternalServerError)
		return
	}
	week, err := sumRange(r.Context(), weekStart)
	if err != nil {
		http.Error(w, "Failed to build stats", http.StatusInternalServerError)
		return
	}
	month, err := sumRange(r.Context(), monthStart)
	if err != nil {
		http.Error(w, "Failed to build stats", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"day": map[string]float64{
			"protein_g": day.protein,
			"fat_g":     day.fat,
			"carbs_g":   day.carbs,
			"calories":  day.calories,
		},
		"week_avg": map[string]float64{
			"protein_g": week.protein / 7,
			"fat_g":     week.fat / 7,
			"carbs_g":   week.carbs / 7,
			"calories":  week.calories / 7,
		},
		"month_avg": map[string]float64{
			"protein_g": month.protein / 30,
			"fat_g":     month.fat / 30,
			"carbs_g":   month.carbs / 30,
			"calories":  month.calories / 30,
		},
	})
}

type catalogItem struct {
	ID       int     `json:"id,omitempty"`
	Title    string  `json:"title"`
	ProteinG float64 `json:"protein_g,omitempty"`
	FatG     float64 `json:"fat_g,omitempty"`
	CarbsG   float64 `json:"carbs_g,omitempty"`
	Calories float64 `json:"calories,omitempty"`
	BaseGrams float64 `json:"base_grams,omitempty"`
}

func (s *Service) searchDishes(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 30
	}

	rows, err := s.pool.Query(
		r.Context(),
		`SELECT id, title, protein_g, fat_g, carbs_g, calories, base_grams
		 FROM sportapp.nutrition_dishes
		 WHERE ($1 = '' OR LOWER(title) LIKE $2)
		 ORDER BY title
		 LIMIT $3`,
		search, "%"+strings.ToLower(search)+"%", limit,
	)
	if err != nil {
		http.Error(w, "Failed to load dishes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]catalogItem, 0, limit)
	for rows.Next() {
		var item catalogItem
		if err := rows.Scan(&item.ID, &item.Title, &item.ProteinG, &item.FatG, &item.CarbsG, &item.Calories, &item.BaseGrams); err != nil {
			http.Error(w, "Failed to scan dishes", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Service) createDish(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var payload struct {
		Title    string   `json:"title"`
		ProteinG *float64 `json:"protein_g"`
		FatG     *float64 `json:"fat_g"`
		CarbsG   *float64 `json:"carbs_g"`
		Calories *float64 `json:"calories"`
		BaseGrams *float64 `json:"base_grams"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	pg := float64(0)
	fg := float64(0)
	cg := float64(0)
	if payload.ProteinG != nil {
		pg = *payload.ProteinG
	}
	if payload.FatG != nil {
		fg = *payload.FatG
	}
	if payload.CarbsG != nil {
		cg = *payload.CarbsG
	}
	cal := caloriesFromMacros(pg, fg, cg)
	if payload.Calories != nil && *payload.Calories > 0 {
		cal = *payload.Calories
	}
	baseGrams := 100.0
	if payload.BaseGrams != nil && *payload.BaseGrams > 0 {
		baseGrams = *payload.BaseGrams
	}
	var item catalogItem
	err = s.pool.QueryRow(
		r.Context(),
		`INSERT INTO sportapp.nutrition_dishes (title, protein_g, fat_g, carbs_g, calories, base_grams, created_by_user_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (title) DO UPDATE
		 SET protein_g = EXCLUDED.protein_g, fat_g = EXCLUDED.fat_g, carbs_g = EXCLUDED.carbs_g, calories = EXCLUDED.calories, base_grams = EXCLUDED.base_grams,
		     created_by_user_id = COALESCE(sportapp.nutrition_dishes.created_by_user_id, EXCLUDED.created_by_user_id),
		     updated_at = NOW()
		 RETURNING id, title, protein_g, fat_g, carbs_g, calories, base_grams`,
		title, pg, fg, cg, cal, baseGrams, userID,
	).Scan(&item.ID, &item.Title, &item.ProteinG, &item.FatG, &item.CarbsG, &item.Calories, &item.BaseGrams)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			http.Error(w, "Dish title already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to save dish", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Service) listMyDishes(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := s.pool.Query(
		r.Context(),
		`SELECT id, title, protein_g, fat_g, carbs_g, calories, base_grams
		 FROM sportapp.nutrition_dishes
		 WHERE created_by_user_id = $1
		 ORDER BY COALESCE(updated_at, created_at) DESC, id DESC`,
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to list dishes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	items := make([]catalogItem, 0, 32)
	for rows.Next() {
		var item catalogItem
		if err := rows.Scan(&item.ID, &item.Title, &item.ProteinG, &item.FatG, &item.CarbsG, &item.Calories, &item.BaseGrams); err != nil {
			http.Error(w, "Failed to scan dishes", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Service) patchMyDish(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
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
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	pg := float64(0)
	fg := float64(0)
	cg := float64(0)
	if payload.ProteinG != nil {
		pg = *payload.ProteinG
	}
	if payload.FatG != nil {
		fg = *payload.FatG
	}
	if payload.CarbsG != nil {
		cg = *payload.CarbsG
	}
	cal := caloriesFromMacros(pg, fg, cg)
	if payload.Calories != nil && *payload.Calories > 0 {
		cal = *payload.Calories
	}
	baseGrams := 100.0
	if payload.BaseGrams != nil && *payload.BaseGrams > 0 {
		baseGrams = *payload.BaseGrams
	}
	var item catalogItem
	err = s.pool.QueryRow(
		r.Context(),
		`UPDATE sportapp.nutrition_dishes
		 SET title = $1, protein_g = $2, fat_g = $3, carbs_g = $4, calories = $5, base_grams = $6, updated_at = NOW()
		 WHERE id = $7 AND created_by_user_id = $8
		 RETURNING id, title, protein_g, fat_g, carbs_g, calories, base_grams`,
		title, pg, fg, cg, cal, baseGrams, id, userID,
	).Scan(&item.ID, &item.Title, &item.ProteinG, &item.FatG, &item.CarbsG, &item.Calories, &item.BaseGrams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Not found or not your dish", http.StatusNotFound)
			return
		}
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			http.Error(w, "Dish title already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to update dish", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Service) deleteMyDish(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	tx, err := s.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "Failed to begin tx", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())
	if _, err := tx.Exec(r.Context(), `UPDATE sportapp.nutrition_entries SET dish_id = NULL WHERE user_id = $1 AND dish_id = $2`, userID, id); err != nil {
		http.Error(w, "Failed to detach entries", http.StatusInternalServerError)
		return
	}
	res, err := tx.Exec(r.Context(), `DELETE FROM sportapp.nutrition_dishes WHERE id = $1 AND created_by_user_id = $2`, id, userID)
	if err != nil {
		http.Error(w, "Failed to delete dish", http.StatusInternalServerError)
		return
	}
	if res.RowsAffected() == 0 {
		http.Error(w, "Not found or not your dish", http.StatusNotFound)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Failed to commit", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) upsertWater(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var payload struct {
		AmountML int    `json:"amount_ml"`
		Day      string `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if payload.AmountML < 0 {
		payload.AmountML = 0
	}
	_, _, defaultYMD := dayBoundsFromQuery(r)
	loggedDay := defaultYMD
	if strings.TrimSpace(payload.Day) != "" {
		if d, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(payload.Day), nutritionTZ); err == nil {
			loggedDay = d.Format("2006-01-02")
		}
	}
	_, err = s.pool.Exec(
		r.Context(),
		`INSERT INTO sportapp.nutrition_water_logs (user_id, logged_on, amount_ml)
		 VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, logged_on) DO UPDATE SET amount_ml = EXCLUDED.amount_ml, updated_at = NOW()`,
		userID, loggedDay, payload.AmountML,
	)
	if err != nil {
		http.Error(w, "Failed to save water", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Service) upsertWeight(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var payload struct {
		WeightKg float64 `json:"weight_kg"`
		Day      string  `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if payload.WeightKg <= 0 {
		http.Error(w, "weight_kg must be > 0", http.StatusBadRequest)
		return
	}
	_, _, defaultYMD := dayBoundsFromQuery(r)
	loggedDay := defaultYMD
	if strings.TrimSpace(payload.Day) != "" {
		if d, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(payload.Day), nutritionTZ); err == nil {
			loggedDay = d.Format("2006-01-02")
		}
	}
	_, err = s.pool.Exec(
		r.Context(),
		`INSERT INTO sportapp.nutrition_weight_logs (user_id, logged_on, weight_kg)
		 VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, logged_on) DO UPDATE SET weight_kg = EXCLUDED.weight_kg, updated_at = NOW()`,
		userID, loggedDay, payload.WeightKg,
	)
	if err != nil {
		http.Error(w, "Failed to save weight", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Service) getDashboard(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	day, dayEnd, dayStr := dayBoundsFromQuery(r)

	type totals struct {
		Protein  float64
		Fat      float64
		Carbs    float64
		Calories float64
	}
	var t totals
	err = s.pool.QueryRow(
		r.Context(),
		`SELECT COALESCE(SUM(protein_g),0), COALESCE(SUM(fat_g),0), COALESCE(SUM(carbs_g),0), COALESCE(SUM(calories),0)
		 FROM sportapp.nutrition_entries
		 WHERE user_id = $1 AND consumed_at >= $2 AND consumed_at < $3`,
		userID, day, dayEnd,
	).Scan(&t.Protein, &t.Fat, &t.Carbs, &t.Calories)
	if err != nil {
		http.Error(w, "Failed to build dashboard", http.StatusInternalServerError)
		return
	}

	var streakDays int
	err = s.pool.QueryRow(
		r.Context(),
		`WITH days AS (
			SELECT d::date AS day
			FROM generate_series(CURRENT_DATE - INTERVAL '180 days', CURRENT_DATE, '1 day') d
		),
		filled AS (
			SELECT day,
			       (
			         EXISTS (SELECT 1 FROM sportapp.nutrition_entries e WHERE e.user_id = $1 AND DATE(e.consumed_at) = day)
			         OR EXISTS (SELECT 1 FROM sportapp.nutrition_water_logs w WHERE w.user_id = $1 AND w.logged_on = day)
			         OR EXISTS (SELECT 1 FROM sportapp.nutrition_weight_logs wl WHERE wl.user_id = $1 AND wl.logged_on = day)
			       ) AS is_filled
			FROM days
		),
		latest_break AS (
			SELECT COALESCE(MAX(day), CURRENT_DATE - INTERVAL '365 days') AS last_break
			FROM filled
			WHERE is_filled = false AND day <= CURRENT_DATE
		)
		SELECT COUNT(*)
		FROM filled f, latest_break b
		WHERE f.is_filled = true AND f.day > b.last_break`,
		userID,
	).Scan(&streakDays)
	if err != nil {
		streakDays = 0
	}

	var goalProtein, goalFat, goalCarbs, goalCalories float64
	err = s.pool.QueryRow(
		r.Context(),
		`SELECT protein_g, fat_g, carbs_g, calories FROM sportapp.nutrition_goals WHERE user_id = $1`,
		userID,
	).Scan(&goalProtein, &goalFat, &goalCarbs, &goalCalories)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			g, syncErr := s.upsertNutritionGoalFromProfile(r.Context(), userID, "", nil)
			if syncErr == nil {
				goalProtein, goalFat, goalCarbs, goalCalories = g.ProteinG, g.FatG, g.CarbsG, g.Calories
			}
		} else {
			http.Error(w, "Failed to load nutrition goal", http.StatusInternalServerError)
			return
		}
	}

	var waterML int
	_ = s.pool.QueryRow(r.Context(), `SELECT amount_ml FROM sportapp.nutrition_water_logs WHERE user_id = $1 AND logged_on = $2`, userID, dayStr).Scan(&waterML)

	var lastWeight *float64
	var lastWeightDay *time.Time
	row := s.pool.QueryRow(r.Context(), `SELECT weight_kg, logged_on::timestamp FROM sportapp.nutrition_weight_logs WHERE user_id = $1 ORDER BY logged_on DESC LIMIT 1`, userID)
	if rowErr := row.Scan(&lastWeight, &lastWeightDay); rowErr != nil && !errors.Is(rowErr, pgx.ErrNoRows) {
		http.Error(w, "Failed to load weight", http.StatusInternalServerError)
		return
	}
	needWeightReminder := true
	if lastWeightDay != nil {
		needWeightReminder = day.Sub(*lastWeightDay) >= 72*time.Hour
	}

	waterGoalKg := 0.0
	if lastWeight != nil && *lastWeight > 0 {
		waterGoalKg = *lastWeight
	} else {
		var profileW *int
		_ = s.pool.QueryRow(r.Context(), `SELECT weight_kg FROM sportapp.profile WHERE user_id = $1`, userID).Scan(&profileW)
		if profileW != nil && *profileW > 0 {
			waterGoalKg = float64(*profileW)
		}
	}
	wMlGoal := dailyWaterGoalMlFromWeight(waterGoalKg)

	writeJSON(w, http.StatusOK, map[string]any{
		"day": dayStr,
		"today": map[string]any{
			"protein_g": t.Protein,
			"fat_g":     t.Fat,
			"carbs_g":   t.Carbs,
			"calories":  t.Calories,
		},
		"goal": map[string]any{
			"calories":           goalCalories,
			"protein_g":          goalProtein,
			"fat_g":              goalFat,
			"carbs_g":            goalCarbs,
			"remaining_calories": goalCalories - t.Calories,
		},
		"water": map[string]any{
			"amount_ml":   waterML,
			"goal_ml":     wMlGoal,
			"goal_liters": math.Round(float64(wMlGoal)/1000.0*100) / 100,
		},
		"weight": map[string]any{
			"last_weight_kg":      lastWeight,
			"last_weight_day":     lastWeightDay,
			"need_weight_reminder": needWeightReminder,
		},
		"streak_days": streakDays,
		"burned_calories": 0.0,
	})
}

func (s *Service) getReports(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("days")))
	if limit <= 0 || limit > 4000 {
		limit = 30
	}

	foodRows, err := s.pool.Query(
		r.Context(),
		`SELECT DATE(consumed_at) as d, COALESCE(SUM(protein_g),0), COALESCE(SUM(fat_g),0), COALESCE(SUM(carbs_g),0), COALESCE(SUM(calories),0)
		 FROM sportapp.nutrition_entries
		 WHERE user_id = $1 AND consumed_at >= NOW() - ($2::int * interval '1 day')
		 GROUP BY d
		 ORDER BY d DESC`,
		userID, limit,
	)
	if err != nil {
		http.Error(w, "Failed to build food report", http.StatusInternalServerError)
		return
	}
	defer foodRows.Close()
	foodSeries := make([]map[string]any, 0, limit)
	for foodRows.Next() {
		var day time.Time
		var p, f, c, kcal float64
		if err := foodRows.Scan(&day, &p, &f, &c, &kcal); err != nil {
			http.Error(w, "Failed to read food report", http.StatusInternalServerError)
			return
		}
		foodSeries = append(foodSeries, map[string]any{
			"day": day.Format("2006-01-02"), "protein_g": p, "fat_g": f, "carbs_g": c, "calories": kcal,
		})
	}

	weightRows, err := s.pool.Query(
		r.Context(),
		`SELECT logged_on, weight_kg
		 FROM sportapp.nutrition_weight_logs
		 WHERE user_id = $1 AND logged_on >= CURRENT_DATE - ($2::int * interval '1 day')
		 ORDER BY logged_on DESC`,
		userID, limit,
	)
	if err != nil {
		http.Error(w, "Failed to build weight report", http.StatusInternalServerError)
		return
	}
	defer weightRows.Close()
	weightSeries := make([]map[string]any, 0, limit)
	for weightRows.Next() {
		var day time.Time
		var weight float64
		if err := weightRows.Scan(&day, &weight); err != nil {
			http.Error(w, "Failed to read weight report", http.StatusInternalServerError)
			return
		}
		weightSeries = append(weightSeries, map[string]any{"day": day.Format("2006-01-02"), "weight_kg": weight})
	}

	waterRows, err := s.pool.Query(
		r.Context(),
		`SELECT logged_on, amount_ml
		 FROM sportapp.nutrition_water_logs
		 WHERE user_id = $1 AND logged_on >= CURRENT_DATE - ($2::int * interval '1 day')
		 ORDER BY logged_on DESC`,
		userID, limit,
	)
	if err != nil {
		http.Error(w, "Failed to build water report", http.StatusInternalServerError)
		return
	}
	defer waterRows.Close()
	waterSeries := make([]map[string]any, 0, limit)
	for waterRows.Next() {
		var day time.Time
		var ml int
		if err := waterRows.Scan(&day, &ml); err != nil {
			http.Error(w, "Failed to read water report", http.StatusInternalServerError)
			return
		}
		waterSeries = append(waterSeries, map[string]any{"day": day.Format("2006-01-02"), "amount_ml": ml})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"days":   limit,
		"food":   foodSeries,
		"weight": weightSeries,
		"water":  waterSeries,
	})
}

