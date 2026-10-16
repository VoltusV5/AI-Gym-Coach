package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	core_errors "sport_app/internal/core/errors"
	nutrition_repo "sport_app/internal/features/nutrition/repository/postgres"
)

// ErrProfileIncomplete indicates that user profile data is insufficient for goal calculation.
var ErrProfileIncomplete = errors.New("profile incomplete")

// Repository defines the data access contract required by NutritionService.
type Repository interface {
	ListEntries(ctx context.Context, userID int, startUTC, endUTC time.Time, limit, offset int) ([]nutrition_repo.NutritionEntry, int, error)
	CreateEntry(ctx context.Context, userID int, entry nutrition_repo.NutritionEntry) (nutrition_repo.NutritionEntry, error)
	UpdateEntry(ctx context.Context, userID int, entryID int, entry nutrition_repo.NutritionEntry) error
	DeleteEntry(ctx context.Context, userID int, entryID int) error

	ListFavorites(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.NutritionFavorite, int, error)
	CreateFavorite(ctx context.Context, userID int, fav nutrition_repo.NutritionFavorite) (int, error)
	DeleteFavorite(ctx context.Context, userID int, favID int) error

	GetGoal(ctx context.Context, userID int) (nutrition_repo.NutritionGoal, error)
	UpsertGoal(ctx context.Context, userID int, goal nutrition_repo.NutritionGoal) (nutrition_repo.NutritionGoal, error)
	GetUserProfile(ctx context.Context, userID int) (nutrition_repo.UserProfileData, error)
	GetLatestWeightKg(ctx context.Context, userID int) (*float64, error)
	GetProfileWeightKg(ctx context.Context, userID int) (*int, error)

	SearchDishes(ctx context.Context, searchQuery string, limit, offset int) ([]nutrition_repo.CatalogDish, int, error)
	GetDishByID(ctx context.Context, dishID int) (nutrition_repo.CatalogDish, error)
	CreateOrUpsertDish(ctx context.Context, userID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error)
	ListMyDishes(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.CatalogDish, int, error)
	UpdateMyDish(ctx context.Context, userID int, dishID int, dish nutrition_repo.CatalogDish) (nutrition_repo.CatalogDish, error)
	DeleteMyDish(ctx context.Context, userID int, dishID int) error

	UpsertWater(ctx context.Context, userID int, loggedDay string, amountML int) error
	UpsertWeight(ctx context.Context, userID int, loggedDay string, weightKg float64) error
	GetDayTotals(ctx context.Context, userID int, startUTC, endUTC time.Time) (nutrition_repo.DailyMacroTotals, error)
	GetStreakDays(ctx context.Context, userID int) (int, error)
	GetWaterLogML(ctx context.Context, userID int, loggedDay string) (int, error)
	GetLatestWeightInfo(ctx context.Context, userID int) (*float64, *time.Time, error)
	GetRangeSums(ctx context.Context, userID int, from time.Time) (nutrition_repo.DailyMacroTotals, error)
	GetAnalyticsFood(ctx context.Context, userID int, daysLimit int) ([]nutrition_repo.FoodDaySummary, error)
	GetAnalyticsWeight(ctx context.Context, userID int, daysLimit int) ([]nutrition_repo.WeightLogSummary, error)
	GetAnalyticsWater(ctx context.Context, userID int, daysLimit int) ([]nutrition_repo.WaterLogSummary, error)
}

// NutritionService implements business logic for nutrition tracking.
type NutritionService struct {
	repo Repository
	tz   *time.Location
}

// NewNutritionService constructs a new NutritionService.
func NewNutritionService(repo Repository) *NutritionService {
	tz, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		tz = time.UTC
	}
	return &NutritionService{
		repo: repo,
		tz:   tz,
	}
}

// CreateEntryInput holds input parameters for creating a food entry.
type CreateEntryInput struct {
	DishID     *int
	Grams      *float64
	MealType   string
	Title      string
	ProteinG   float64
	FatG       float64
	CarbsG     float64
	Calories   *float64
	ConsumedAt *string
	Day        string
}

// UpdateEntryInput holds input parameters for updating a food entry.
type UpdateEntryInput struct {
	DishID     *int
	Grams      *float64
	MealType   string
	Title      string
	ProteinG   float64
	FatG       float64
	CarbsG     float64
	Calories   *float64
	ConsumedAt *string
	Day        string
}

// DashboardResult aggregates all daily nutrition tracker data.
type DashboardResult struct {
	Day            string         `json:"day"`
	Today          map[string]any `json:"today"`
	Goal           map[string]any `json:"goal"`
	Water          map[string]any `json:"water"`
	Weight         map[string]any `json:"weight"`
	StreakDays     int            `json:"streak_days"`
	BurnedCalories float64        `json:"burned_calories"`
}

// StatsResult aggregates macronutrient summaries across day, week and month.
type StatsResult struct {
	Day      map[string]float64 `json:"day"`
	WeekAvg  map[string]float64 `json:"week_avg"`
	MonthAvg map[string]float64 `json:"month_avg"`
}

// AnalyticsResult contains historical charts for food, weight, and water.
type AnalyticsResult struct {
	Days   int                               `json:"days"`
	Food   []nutrition_repo.FoodDaySummary   `json:"food"`
	Weight []nutrition_repo.WeightLogSummary `json:"weight"`
	Water  []nutrition_repo.WaterLogSummary  `json:"water"`
}

// CaloriesFromMacros calculates total calories from protein, fat, and carbohydrate grams.
func CaloriesFromMacros(p, f, c float64) float64 {
	return p*4 + f*9 + c*4
}

// ScaleFrom100 scales nutrition values per 100g to specified gram quantity.
func ScaleFrom100(valuePer100, grams float64) float64 {
	if grams <= 0 {
		grams = 100
	}
	return (valuePer100 / 100.0) * grams
}

// MealTypeOrDefault normalizes meal type or defaults to "snack".
func MealTypeOrDefault(v string) string {
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

// DailyWaterGoalMlFromWeight computes target daily water intake in milliliters based on body weight.
func DailyWaterGoalMlFromWeight(weightKg float64) int {
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

// ActivityMultiplier returns the physical activity coefficient.
func ActivityMultiplier(level string) float64 {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "сидячий и малоподвижный":
		return 1.2
	case "лёгкая активность (физические нагрузки 1-3 раза в неделю)",
		"легкая активность (физические нагрузки 1-3 раза в неделю)":
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

// TargetAdjustment adjusts calorie target according to fitness goal.
func TargetAdjustment(target string, weightKg float64) float64 {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "lose", "сбросить вес", "сжечь жир":
		return -2 * weightKg
	case "gain", "набрать мышцы", "набрать мышцы и сжечь жир":
		return 2 * weightKg
	default:
		return 0
	}
}

// TargetDeltaAdjustment adjusts calories for target weight delta.
func TargetDeltaAdjustment(deltaKg float64) float64 {
	return deltaKg * 220
}

// CalculateGoalFromProfile computes daily nutrition target macros using Mifflin-St Jeor formula.
func CalculateGoalFromProfile(age int, heightCm, weightKg float64, gender, activity, target string) nutrition_repo.NutritionGoal {
	base := (10*weightKg + 6.25*heightCm - 5*float64(age))
	if strings.EqualFold(strings.TrimSpace(gender), "Женщина") || strings.EqualFold(strings.TrimSpace(gender), "female") {
		base -= 161
	} else {
		base += 5
	}
	calories := base*ActivityMultiplier(activity) + TargetAdjustment(target, weightKg)
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
	return nutrition_repo.NutritionGoal{
		ProteinG: protein,
		FatG:     fat,
		CarbsG:   carbs,
		Calories: calories,
	}
}

// DayBoundsFromQuery computes start and end UTC bounds for the given day string.
func (s *NutritionService) DayBoundsFromQuery(dayRaw string) (time.Time, time.Time, string) {
	now := time.Now().In(s.tz)
	var dayStart time.Time
	if dayRaw == "" {
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.tz)
	} else if d, err := time.ParseInLocation("2006-01-02", dayRaw, s.tz); err == nil {
		dayStart = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, s.tz)
	} else {
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.tz)
	}
	end := dayStart.Add(24 * time.Hour)
	return dayStart.UTC(), end.UTC(), dayStart.Format("2006-01-02")
}

func (s *NutritionService) parseConsumedAt(consumedAt *string, day string) (time.Time, error) {
	if strings.TrimSpace(day) != "" {
		d, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(day), s.tz)
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, s.tz).UTC(), nil
	}
	if consumedAt == nil || strings.TrimSpace(*consumedAt) == "" {
		return time.Now().UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*consumedAt))
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// ListEntries returns logged entries for a specific day with optional pagination.
func (s *NutritionService) ListEntries(
	ctx context.Context,
	userID int,
	dayRaw string,
	limit int,
	offset int,
) ([]nutrition_repo.NutritionEntry, int, string, error) {
	startUTC, endUTC, dayYMD := s.DayBoundsFromQuery(dayRaw)
	items, total, err := s.repo.ListEntries(ctx, userID, startUTC, endUTC, limit, offset)
	if err != nil {
		return nil, 0, "", err
	}
	return items, total, dayYMD, nil
}

// CreateEntry adds a new food log entry with automatic macro scaling and dish association.
func (s *NutritionService) CreateEntry(
	ctx context.Context,
	userID int,
	inp CreateEntryInput,
) (nutrition_repo.NutritionEntry, error) {
	consumedAt, err := s.parseConsumedAt(inp.ConsumedAt, inp.Day)
	if err != nil {
		return nutrition_repo.NutritionEntry{}, fmt.Errorf("invalid consumed_at or day: %w", core_errors.ErrInvalidArgument)
	}

	mealType := MealTypeOrDefault(inp.MealType)
	grams := 100.0
	if inp.Grams != nil && *inp.Grams > 0 {
		grams = *inp.Grams
	}

	title := strings.TrimSpace(inp.Title)
	if title == "" {
		title = "Без названия"
	}

	proteinG := inp.ProteinG
	fatG := inp.FatG
	carbsG := inp.CarbsG

	var dishID *int
	if inp.DishID != nil && *inp.DishID > 0 {
		dish, err := s.repo.GetDishByID(ctx, *inp.DishID)
		if err == nil {
			dishID = inp.DishID
			proteinG = ScaleFrom100(dish.ProteinG, grams)
			fatG = ScaleFrom100(dish.FatG, grams)
			carbsG = ScaleFrom100(dish.CarbsG, grams)
			if inp.Calories == nil || *inp.Calories <= 0 {
				cal := ScaleFrom100(dish.Calories, grams)
				inp.Calories = &cal
			}
		} else {
			proteinG = ScaleFrom100(proteinG, grams)
			fatG = ScaleFrom100(fatG, grams)
			carbsG = ScaleFrom100(carbsG, grams)
			if inp.Calories != nil && *inp.Calories > 0 {
				cal := ScaleFrom100(*inp.Calories, grams)
				inp.Calories = &cal
			} else {
				cal := CaloriesFromMacros(proteinG, fatG, carbsG)
				inp.Calories = &cal
			}
			dishID = nil
		}
	} else {
		perP, perF, perC := proteinG, fatG, carbsG
		var perK float64
		if inp.Calories != nil && *inp.Calories > 0 {
			perK = *inp.Calories
		} else {
			perK = CaloriesFromMacros(perP, perF, perC)
		}

		createdDish, err := s.repo.CreateOrUpsertDish(ctx, userID, nutrition_repo.CatalogDish{
			Title:     title,
			ProteinG:  perP,
			FatG:      perF,
			CarbsG:    perC,
			Calories:  CaloriesFromMacros(perP, perF, perC),
			BaseGrams: 100,
		})
		if err == nil {
			dishID = &createdDish.ID
			proteinG = ScaleFrom100(perP, grams)
			fatG = ScaleFrom100(perF, grams)
			carbsG = ScaleFrom100(perC, grams)
			cal := ScaleFrom100(perK, grams)
			inp.Calories = &cal
		}
	}

	calories := CaloriesFromMacros(proteinG, fatG, carbsG)
	if inp.Calories != nil && *inp.Calories > 0 {
		calories = *inp.Calories
	}

	entry := nutrition_repo.NutritionEntry{
		DishID:     dishID,
		Grams:      grams,
		MealType:   mealType,
		Title:      title,
		ProteinG:   proteinG,
		FatG:       fatG,
		CarbsG:     carbsG,
		Calories:   calories,
		ConsumedAt: consumedAt,
	}

	return s.repo.CreateEntry(ctx, userID, entry)
}

// UpdateEntry modifies an existing food log entry.
func (s *NutritionService) UpdateEntry(
	ctx context.Context,
	userID int,
	entryID int,
	inp UpdateEntryInput,
) error {
	consumedAt, err := s.parseConsumedAt(inp.ConsumedAt, inp.Day)
	if err != nil {
		return fmt.Errorf("invalid consumed_at or day: %w", core_errors.ErrInvalidArgument)
	}

	mealType := MealTypeOrDefault(inp.MealType)
	grams := 100.0
	if inp.Grams != nil && *inp.Grams > 0 {
		grams = *inp.Grams
	}

	title := strings.TrimSpace(inp.Title)
	if title == "" {
		title = "Без названия"
	}

	proteinG := inp.ProteinG
	fatG := inp.FatG
	carbsG := inp.CarbsG

	if inp.DishID != nil && *inp.DishID > 0 {
		dish, err := s.repo.GetDishByID(ctx, *inp.DishID)
		if err == nil {
			proteinG = ScaleFrom100(dish.ProteinG, grams)
			fatG = ScaleFrom100(dish.FatG, grams)
			carbsG = ScaleFrom100(dish.CarbsG, grams)
			if inp.Calories == nil || *inp.Calories <= 0 {
				cal := ScaleFrom100(dish.Calories, grams)
				inp.Calories = &cal
			}
		} else {
			proteinG = ScaleFrom100(proteinG, grams)
			fatG = ScaleFrom100(fatG, grams)
			carbsG = ScaleFrom100(carbsG, grams)
			if inp.Calories != nil && *inp.Calories > 0 {
				cal := ScaleFrom100(*inp.Calories, grams)
				inp.Calories = &cal
			} else {
				cal := CaloriesFromMacros(proteinG, fatG, carbsG)
				inp.Calories = &cal
			}
		}
	}

	calories := CaloriesFromMacros(proteinG, fatG, carbsG)
	if inp.Calories != nil && *inp.Calories > 0 {
		calories = *inp.Calories
	}

	entry := nutrition_repo.NutritionEntry{
		Grams:      grams,
		MealType:   mealType,
		Title:      title,
		ProteinG:   proteinG,
		FatG:       fatG,
		CarbsG:     carbsG,
		Calories:   calories,
		ConsumedAt: consumedAt,
	}

	return s.repo.UpdateEntry(ctx, userID, entryID, entry)
}

// DeleteEntry deletes an entry owned by the user.
func (s *NutritionService) DeleteEntry(ctx context.Context, userID int, entryID int) error {
	return s.repo.DeleteEntry(ctx, userID, entryID)
}

// ListFavorites returns favorite foods for a user.
func (s *NutritionService) ListFavorites(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.NutritionFavorite, int, error) {
	return s.repo.ListFavorites(ctx, userID, limit, offset)
}

// CreateFavorite saves a new favorite food item.
func (s *NutritionService) CreateFavorite(
	ctx context.Context,
	userID int,
	title string,
	proteinG, fatG, carbsG float64,
	unitType string,
) (int, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Без названия"
	}
	if strings.TrimSpace(unitType) == "" {
		unitType = "gram"
	}
	return s.repo.CreateFavorite(ctx, userID, nutrition_repo.NutritionFavorite{
		Title:    title,
		ProteinG: proteinG,
		FatG:     fatG,
		CarbsG:   carbsG,
		UnitType: unitType,
	})
}

// DeleteFavorite removes a favorite food item.
func (s *NutritionService) DeleteFavorite(ctx context.Context, userID int, favID int) error {
	return s.repo.DeleteFavorite(ctx, userID, favID)
}

// GetGoal returns the nutrition goal, computing and persisting it from profile if not yet set.
func (s *NutritionService) GetGoal(ctx context.Context, userID int) (nutrition_repo.NutritionGoal, error) {
	g, err := s.repo.GetGoal(ctx, userID)
	if err == nil {
		return g, nil
	}
	if !errors.Is(err, core_errors.ErrNotFound) {
		return nutrition_repo.NutritionGoal{}, err
	}

	synced, syncErr := s.UpsertGoalFromProfile(ctx, userID, "", nil)
	if syncErr != nil {
		return nutrition_repo.NutritionGoal{}, core_errors.ErrNotFound
	}
	return synced, nil
}

// UpsertGoalFromProfile calculates and persists target goals from user profile.
func (s *NutritionService) UpsertGoalFromProfile(
	ctx context.Context,
	userID int,
	reqTarget string,
	targetDeltaKg *float64,
) (nutrition_repo.NutritionGoal, error) {
	p, err := s.repo.GetUserProfile(ctx, userID)
	if err != nil {
		return nutrition_repo.NutritionGoal{}, err
	}

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
		lw, qerr := s.repo.GetLatestWeightKg(ctx, userID)
		if qerr == nil && lw != nil && *lw > 30 && *lw < 400 {
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

	newGoal := CalculateGoalFromProfile(age, heightCm, weightKg, gender, activity, goal)
	if targetDeltaKg != nil && math.Abs(*targetDeltaKg) > 1e-6 {
		newGoal.Calories += TargetDeltaAdjustment(*targetDeltaKg)
		if newGoal.Calories < 1200 {
			newGoal.Calories = 1200
		}
		newGoal.FatG = (newGoal.Calories * 0.28) / 9
		newGoal.CarbsG = (newGoal.Calories - newGoal.ProteinG*4 - newGoal.FatG*9) / 4
		if newGoal.CarbsG < 0 {
			newGoal.CarbsG = 0
		}
	}

	return s.repo.UpsertGoal(ctx, userID, newGoal)
}

// RecalculateGoal updates goal target based on profile and optional target delta.
func (s *NutritionService) RecalculateGoal(
	ctx context.Context,
	userID int,
	target string,
	deltaKg *float64,
) (nutrition_repo.NutritionGoal, error) {
	return s.UpsertGoalFromProfile(ctx, userID, target, deltaKg)
}

// GetStats compiles aggregate macronutrient averages for day, week, and month.
func (s *NutritionService) GetStats(ctx context.Context, userID int) (StatsResult, error) {
	now := time.Now().In(s.tz)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.tz)
	weekStart := dayStart.AddDate(0, 0, -7)
	monthStart := dayStart.AddDate(0, 0, -30)

	dayEnd := dayStart.Add(24 * time.Hour)
	daySums, err := s.repo.GetDayTotals(ctx, userID, dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		return StatsResult{}, err
	}

	weekSums, err := s.repo.GetRangeSums(ctx, userID, weekStart.UTC())
	if err != nil {
		return StatsResult{}, err
	}

	monthSums, err := s.repo.GetRangeSums(ctx, userID, monthStart.UTC())
	if err != nil {
		return StatsResult{}, err
	}

	return StatsResult{
		Day: map[string]float64{
			"protein_g": daySums.Protein,
			"fat_g":     daySums.Fat,
			"carbs_g":   daySums.Carbs,
			"calories":  daySums.Calories,
		},
		WeekAvg: map[string]float64{
			"protein_g": weekSums.Protein / 7,
			"fat_g":     weekSums.Fat / 7,
			"carbs_g":   weekSums.Carbs / 7,
			"calories":  weekSums.Calories / 7,
		},
		MonthAvg: map[string]float64{
			"protein_g": monthSums.Protein / 30,
			"fat_g":     monthSums.Fat / 30,
			"carbs_g":   monthSums.Carbs / 30,
			"calories":  monthSums.Calories / 30,
		},
	}, nil
}

// SearchDishes returns catalog dishes matching a search string.
func (s *NutritionService) SearchDishes(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]nutrition_repo.CatalogDish, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 30
	}
	return s.repo.SearchDishes(ctx, strings.TrimSpace(query), limit, offset)
}

// CreateDish creates or updates a custom dish in the shared catalog.
func (s *NutritionService) CreateDish(
	ctx context.Context,
	userID int,
	dish nutrition_repo.CatalogDish,
) (nutrition_repo.CatalogDish, error) {
	dish.Title = strings.TrimSpace(dish.Title)
	if dish.Title == "" {
		return nutrition_repo.CatalogDish{}, fmt.Errorf("dish title is required: %w", core_errors.ErrInvalidArgument)
	}
	if dish.BaseGrams <= 0 {
		dish.BaseGrams = 100
	}
	if dish.Calories <= 0 {
		dish.Calories = CaloriesFromMacros(dish.ProteinG, dish.FatG, dish.CarbsG)
	}
	return s.repo.CreateOrUpsertDish(ctx, userID, dish)
}

// ListMyDishes returns dishes created by the user.
func (s *NutritionService) ListMyDishes(
	ctx context.Context,
	userID int,
	limit, offset int,
) ([]nutrition_repo.CatalogDish, int, error) {
	return s.repo.ListMyDishes(ctx, userID, limit, offset)
}

// UpdateMyDish updates an existing custom dish created by the user.
func (s *NutritionService) UpdateMyDish(
	ctx context.Context,
	userID int,
	dishID int,
	dish nutrition_repo.CatalogDish,
) (nutrition_repo.CatalogDish, error) {
	dish.Title = strings.TrimSpace(dish.Title)
	if dish.Title == "" {
		return nutrition_repo.CatalogDish{}, fmt.Errorf("dish title is required: %w", core_errors.ErrInvalidArgument)
	}
	if dish.BaseGrams <= 0 {
		dish.BaseGrams = 100
	}
	if dish.Calories <= 0 {
		dish.Calories = CaloriesFromMacros(dish.ProteinG, dish.FatG, dish.CarbsG)
	}
	return s.repo.UpdateMyDish(ctx, userID, dishID, dish)
}

// DeleteMyDish deletes a custom dish owned by the user.
func (s *NutritionService) DeleteMyDish(ctx context.Context, userID int, dishID int) error {
	return s.repo.DeleteMyDish(ctx, userID, dishID)
}

// UpsertWater records water consumption.
func (s *NutritionService) UpsertWater(ctx context.Context, userID int, dayRaw string, amountML int) error {
	if amountML < 0 {
		amountML = 0
	}
	_, _, defaultYMD := s.DayBoundsFromQuery(dayRaw)
	return s.repo.UpsertWater(ctx, userID, defaultYMD, amountML)
}

// UpsertWeight records daily body weight.
func (s *NutritionService) UpsertWeight(ctx context.Context, userID int, dayRaw string, weightKg float64) error {
	if weightKg <= 0 {
		return fmt.Errorf("weight_kg must be > 0: %w", core_errors.ErrInvalidArgument)
	}
	_, _, defaultYMD := s.DayBoundsFromQuery(dayRaw)
	return s.repo.UpsertWeight(ctx, userID, defaultYMD, weightKg)
}

// GetDashboard aggregates daily summaries, goals, water, weight and streak.
func (s *NutritionService) GetDashboard(
	ctx context.Context,
	userID int,
	dayRaw string,
) (DashboardResult, error) {
	dayStart, dayEnd, dayStr := s.DayBoundsFromQuery(dayRaw)

	totals, err := s.repo.GetDayTotals(ctx, userID, dayStart, dayEnd)
	if err != nil {
		return DashboardResult{}, err
	}

	streakDays, _ := s.repo.GetStreakDays(ctx, userID)

	goal, err := s.GetGoal(ctx, userID)
	if err != nil && !errors.Is(err, core_errors.ErrNotFound) {
		return DashboardResult{}, err
	}

	waterML, _ := s.repo.GetWaterLogML(ctx, userID, dayStr)
	lastWeight, lastWeightDay, _ := s.repo.GetLatestWeightInfo(ctx, userID)

	needWeightReminder := true
	if lastWeightDay != nil {
		needWeightReminder = dayStart.Sub(*lastWeightDay) >= 72*time.Hour
	}

	waterGoalKg := 0.0
	if lastWeight != nil && *lastWeight > 0 {
		waterGoalKg = *lastWeight
	} else {
		pw, _ := s.repo.GetProfileWeightKg(ctx, userID)
		if pw != nil && *pw > 0 {
			waterGoalKg = float64(*pw)
		}
	}
	wMlGoal := DailyWaterGoalMlFromWeight(waterGoalKg)

	return DashboardResult{
		Day: dayStr,
		Today: map[string]any{
			"protein_g": totals.Protein,
			"fat_g":     totals.Fat,
			"carbs_g":   totals.Carbs,
			"calories":  totals.Calories,
		},
		Goal: map[string]any{
			"calories":           goal.Calories,
			"protein_g":          goal.ProteinG,
			"fat_g":              goal.FatG,
			"carbs_g":            goal.CarbsG,
			"remaining_calories": goal.Calories - totals.Calories,
		},
		Water: map[string]any{
			"amount_ml":   waterML,
			"goal_ml":     wMlGoal,
			"goal_liters": math.Round(float64(wMlGoal)/1000.0*100) / 100,
		},
		Weight: map[string]any{
			"last_weight_kg":       lastWeight,
			"last_weight_day":      lastWeightDay,
			"need_weight_reminder": needWeightReminder,
		},
		StreakDays:     streakDays,
		BurnedCalories: 0.0,
	}, nil
}

// GetAnalytics compiles historical intake, weight, and water charts.
func (s *NutritionService) GetAnalytics(ctx context.Context, userID int, days int) (AnalyticsResult, error) {
	if days <= 0 || days > 4000 {
		days = 30
	}

	food, err := s.repo.GetAnalyticsFood(ctx, userID, days)
	if err != nil {
		return AnalyticsResult{}, err
	}

	weight, err := s.repo.GetAnalyticsWeight(ctx, userID, days)
	if err != nil {
		return AnalyticsResult{}, err
	}

	water, err := s.repo.GetAnalyticsWater(ctx, userID, days)
	if err != nil {
		return AnalyticsResult{}, err
	}

	return AnalyticsResult{
		Days:   days,
		Food:   food,
		Weight: weight,
		Water:  water,
	}, nil
}
