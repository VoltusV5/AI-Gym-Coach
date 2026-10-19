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

// ErrProfileIncomplete означает, что данных профиля недостаточно для расчёта целей питания.
var ErrProfileIncomplete = errors.New("profile incomplete")

// Repository задаёт доступ к данным питания.
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

// NutritionService реализует бизнес-правила дневника питания.
type NutritionService struct {
	repo Repository
	tz   *time.Location
}

// NewNutritionService создаёт сервис питания с репозиторием и часовым поясом дневника.
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

// CreateEntryInput содержит параметры создания записи питания.
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

// UpdateEntryInput содержит параметры изменения записи питания.
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

// DashboardResult объединяет суточные показатели питания, воды, веса и целей.
type DashboardResult struct {
	Day            string         `json:"day"`
	Today          map[string]any `json:"today"`
	Goal           map[string]any `json:"goal"`
	Water          map[string]any `json:"water"`
	Weight         map[string]any `json:"weight"`
	StreakDays     int            `json:"streak_days"`
	BurnedCalories float64        `json:"burned_calories"`
}

// StatsResult содержит суммы и средние значения БЖУ за день, неделю и месяц.
type StatsResult struct {
	Day      map[string]float64 `json:"day"`
	WeekAvg  map[string]float64 `json:"week_avg"`
	MonthAvg map[string]float64 `json:"month_avg"`
}

// AnalyticsResult содержит историю питания, веса и воды для графиков.
type AnalyticsResult struct {
	Days   int                               `json:"days"`
	Food   []nutrition_repo.FoodDaySummary   `json:"food"`
	Weight []nutrition_repo.WeightLogSummary `json:"weight"`
	Water  []nutrition_repo.WaterLogSummary  `json:"water"`
}

// CaloriesFromMacros рассчитывает калорийность по массе белков p, жиров f и углеводов c в граммах.
func CaloriesFromMacros(p, f, c float64) float64 {
	return p*proteinCaloriesPerGram + f*fatCaloriesPerGram + c*carbCaloriesPerGram
}

// ScaleFrom100 пересчитывает valuePer100 с порции 100 г на grams; для неположительной массы использует 100 г.
func ScaleFrom100(valuePer100, grams float64) float64 {
	if grams <= 0 {
		grams = referencePortionGrams
	}
	return (valuePer100 / referencePortionGrams) * grams
}

// MealTypeOrDefault нормализует тип приёма пищи; неизвестное значение заменяет на snack.
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

// DailyWaterGoalMlFromWeight рассчитывает суточную цель воды в миллилитрах по массе тела weightKg.
func DailyWaterGoalMlFromWeight(weightKg float64) int {
	if weightKg <= 0 {
		return defaultWaterGoalML
	}
	ml := int(math.Round(weightKg * waterMLPerKg))
	if ml < minWaterGoalML {
		ml = minWaterGoalML
	}
	if ml > maxWaterGoalML {
		ml = maxWaterGoalML
	}
	return ml
}

// ActivityMultiplier возвращает коэффициент физической активности.
func ActivityMultiplier(level string) float64 {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "сидячий и малоподвижный":
		return sedentaryActivityMultiplier
	case "лёгкая активность (физические нагрузки 1-3 раза в неделю)",
		"легкая активность (физические нагрузки 1-3 раза в неделю)":
		return lightActivityMultiplier
	case "средняя активность (физические нагрузки 3-5 раза в неделю)":
		return moderateActivityMultiplier
	case "высокая активность (физические нагрузки 6-7 раз в неделю)":
		return highActivityMultiplier
	case "очень высокая активность (постоянно высокая физическая нагрузка)":
		return veryHighActivityMultiplier
	default:
		return lightActivityMultiplier
	}
}

// TargetAdjustment рассчитывает поправку калорийности для цели и массы тела.
func TargetAdjustment(target string, weightKg float64) float64 {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "lose", "сбросить вес", "сжечь жир":
		return -targetCaloriesPerKg * weightKg
	case "gain", "набрать мышцы", "набрать мышцы и сжечь жир":
		return targetCaloriesPerKg * weightKg
	default:
		return 0
	}
}

// TargetDeltaAdjustment рассчитывает поправку калорийности по целевому изменению массы.
func TargetDeltaAdjustment(deltaKg float64) float64 {
	return deltaKg * targetDeltaCaloriesPerKg
}

// CalculateGoalFromProfile рассчитывает суточные калории и БЖУ по формуле Миффлина — Сан Жеора, активности и цели.
func CalculateGoalFromProfile(age int, heightCm, weightKg float64, gender, activity, target string) nutrition_repo.NutritionGoal {
	base := (bmrWeightCoefficient*weightKg + bmrHeightCoefficient*heightCm - bmrAgeCoefficient*float64(age))
	if strings.EqualFold(strings.TrimSpace(gender), "Женщина") || strings.EqualFold(strings.TrimSpace(gender), "female") {
		base -= bmrFemaleOffset
	} else {
		base += bmrMaleOffset
	}
	calories := base*ActivityMultiplier(activity) + TargetAdjustment(target, weightKg)
	if calories < minimumDailyCalories {
		calories = minimumDailyCalories
	}

	var protein float64
	var fat float64
	if strings.Contains(strings.ToLower(target), "сброс") || strings.Contains(strings.ToLower(target), "сжечь") {
		protein = weightKg * weightLossProteinPerKg
		fat = (calories * weightLossFatFraction) / fatCaloriesPerGram
	} else if strings.Contains(strings.ToLower(target), "набрать") {
		protein = weightKg * muscleGainProteinPerKg
		fat = (calories * muscleGainFatFraction) / fatCaloriesPerGram
	} else {
		protein = weightKg * maintenanceProteinPerKg
		fat = (calories * maintenanceFatFraction) / fatCaloriesPerGram
	}
	carbs := (calories - protein*proteinCaloriesPerGram - fat*fatCaloriesPerGram) / carbCaloriesPerGram
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

// DayBoundsFromQuery возвращает UTC-границы дня и его дату в часовом поясе дневника; при пустой или неверной дате использует сегодня.
func (s *NutritionService) DayBoundsFromQuery(dayRaw string) (time.Time, time.Time, string) {
	now := time.Now().In(s.tz)
	var dayStart time.Time
	if dayRaw == "" {
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.tz)
	} else if d, err := time.ParseInLocation(time.DateOnly, dayRaw, s.tz); err == nil {
		dayStart = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, s.tz)
	} else {
		dayStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.tz)
	}
	end := dayStart.Add(nutritionDayDuration)
	return dayStart.UTC(), end.UTC(), dayStart.Format(time.DateOnly)
}

func (s *NutritionService) parseConsumedAt(consumedAt *string, day string) (time.Time, error) {
	if strings.TrimSpace(day) != "" {
		d, err := time.ParseInLocation(time.DateOnly, strings.TrimSpace(day), s.tz)
		if err != nil {
			return time.Time{}, err
		}
		return time.Date(d.Year(), d.Month(), d.Day(), defaultMealHour, 0, 0, 0, s.tz).UTC(), nil
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

// ListEntries возвращает записи питания пользователя за день со стабильной сортировкой и необязательной пагинацией.
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

// CreateEntry сохраняет запись питания пользователя с пересчётом БЖУ по массе порции и привязкой к блюду. Возвращает запись либо ошибку.
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
	grams := float64(referencePortionGrams)
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
			BaseGrams: referencePortionGrams,
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

// UpdateEntry изменяет принадлежащую пользователю запись питания. Возвращает запись либо ошибку.
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
	grams := float64(referencePortionGrams)
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

// DeleteEntry удаляет запись питания только её владельца; при отсутствии записи возвращает ErrNotFound.
func (s *NutritionService) DeleteEntry(ctx context.Context, userID int, entryID int) error {
	return s.repo.DeleteEntry(ctx, userID, entryID)
}

// ListFavorites возвращает избранные продукты пользователя.
func (s *NutritionService) ListFavorites(ctx context.Context, userID int, limit, offset int) ([]nutrition_repo.NutritionFavorite, int, error) {
	return s.repo.ListFavorites(ctx, userID, limit, offset)
}

// CreateFavorite сохраняет продукт в избранном пользователя.
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

// DeleteFavorite удаляет продукт из избранного только его владельца.
func (s *NutritionService) DeleteFavorite(ctx context.Context, userID int, favID int) error {
	return s.repo.DeleteFavorite(ctx, userID, favID)
}

// GetGoal возвращает цель питания пользователя; сервис рассчитывает и сохраняет её по профилю при отсутствии.
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

// UpsertGoalFromProfile рассчитывает и сохраняет цели питания по профилю и необязательному целевому изменению массы.
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

	age := defaultProfileAge
	if p.Age != nil && *p.Age > 0 && *p.Age < maximumProfileAge {
		age = *p.Age
	}
	heightCm := float64(defaultProfileHeightCM)
	if p.HeightCm != nil && *p.HeightCm > minimumProfileHeightCM && *p.HeightCm < maximumProfileHeightCM {
		heightCm = float64(*p.HeightCm)
	}
	weightKg := float64(defaultProfileWeightKG)
	if p.WeightKg != nil && *p.WeightKg > minimumProfileWeightKG && *p.WeightKg < maximumProfileWeightKG {
		weightKg = float64(*p.WeightKg)
	} else {
		lw, qerr := s.repo.GetLatestWeightKg(ctx, userID)
		if qerr == nil && lw != nil && *lw > minimumProfileWeightKG && *lw < maximumProfileWeightKG {
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
	if targetDeltaKg != nil && math.Abs(*targetDeltaKg) > targetDeltaEpsilon {
		newGoal.Calories += TargetDeltaAdjustment(*targetDeltaKg)
		if newGoal.Calories < minimumDailyCalories {
			newGoal.Calories = minimumDailyCalories
		}
		newGoal.FatG = (newGoal.Calories * recalculatedFatFraction) / fatCaloriesPerGram
		newGoal.CarbsG = (newGoal.Calories - newGoal.ProteinG*proteinCaloriesPerGram - newGoal.FatG*fatCaloriesPerGram) / carbCaloriesPerGram
		if newGoal.CarbsG < 0 {
			newGoal.CarbsG = 0
		}
	}

	return s.repo.UpsertGoal(ctx, userID, newGoal)
}

// RecalculateGoal пересчитывает цели питания по профилю и необязательному изменению массы.
func (s *NutritionService) RecalculateGoal(
	ctx context.Context,
	userID int,
	target string,
	deltaKg *float64,
) (nutrition_repo.NutritionGoal, error) {
	return s.UpsertGoalFromProfile(ctx, userID, target, deltaKg)
}

// GetStats возвращает суммарные и средние значения БЖУ и калорий за день, неделю и месяц.
func (s *NutritionService) GetStats(ctx context.Context, userID int) (StatsResult, error) {
	now := time.Now().In(s.tz)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.tz)
	weekStart := dayStart.AddDate(0, 0, -statsWeekDays)
	monthStart := dayStart.AddDate(0, 0, -statsMonthDays)

	dayEnd := dayStart.Add(nutritionDayDuration)
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
			"protein_g": weekSums.Protein / statsWeekDays,
			"fat_g":     weekSums.Fat / statsWeekDays,
			"carbs_g":   weekSums.Carbs / statsWeekDays,
			"calories":  weekSums.Calories / statsWeekDays,
		},
		MonthAvg: map[string]float64{
			"protein_g": monthSums.Protein / statsMonthDays,
			"fat_g":     monthSums.Fat / statsMonthDays,
			"carbs_g":   monthSums.Carbs / statsMonthDays,
			"calories":  monthSums.Calories / statsMonthDays,
		},
	}, nil
}

// SearchDishes ищет блюда по подстроке названия со стабильной сортировкой.
func (s *NutritionService) SearchDishes(
	ctx context.Context,
	query string,
	limit, offset int,
) ([]nutrition_repo.CatalogDish, int, error) {
	if limit <= 0 || limit > maxDishSearchLimit {
		limit = DefaultDishSearchLimit
	}
	return s.repo.SearchDishes(ctx, strings.TrimSpace(query), limit, offset)
}

// CreateDish создаёт или обновляет пользовательское блюдо в каталоге.
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
		dish.BaseGrams = referencePortionGrams
	}
	if dish.Calories <= 0 {
		dish.Calories = CaloriesFromMacros(dish.ProteinG, dish.FatG, dish.CarbsG)
	}
	return s.repo.CreateOrUpsertDish(ctx, userID, dish)
}

// ListMyDishes возвращает блюда, созданные текущим пользователем.
func (s *NutritionService) ListMyDishes(
	ctx context.Context,
	userID int,
	limit, offset int,
) ([]nutrition_repo.CatalogDish, int, error) {
	return s.repo.ListMyDishes(ctx, userID, limit, offset)
}

// UpdateMyDish изменяет блюдо только для его автора.
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
		dish.BaseGrams = referencePortionGrams
	}
	if dish.Calories <= 0 {
		dish.Calories = CaloriesFromMacros(dish.ProteinG, dish.FatG, dish.CarbsG)
	}
	return s.repo.UpdateMyDish(ctx, userID, dishID, dish)
}

// DeleteMyDish удаляет блюдо только для его автора.
func (s *NutritionService) DeleteMyDish(ctx context.Context, userID int, dishID int) error {
	return s.repo.DeleteMyDish(ctx, userID, dishID)
}

// UpsertWater сохраняет объём выпитой воды пользователя за указанную дату.
func (s *NutritionService) UpsertWater(ctx context.Context, userID int, dayRaw string, amountML int) error {
	if amountML < 0 {
		amountML = 0
	}
	_, _, defaultYMD := s.DayBoundsFromQuery(dayRaw)
	return s.repo.UpsertWater(ctx, userID, defaultYMD, amountML)
}

// UpsertWeight сохраняет массу тела пользователя за указанную дату.
func (s *NutritionService) UpsertWeight(ctx context.Context, userID int, dayRaw string, weightKg float64) error {
	if weightKg <= 0 {
		return fmt.Errorf("weight_kg must be > 0: %w", core_errors.ErrInvalidArgument)
	}
	_, _, defaultYMD := s.DayBoundsFromQuery(dayRaw)
	return s.repo.UpsertWeight(ctx, userID, defaultYMD, weightKg)
}

// GetDashboard объединяет суточные показатели питания, цели, воду, вес и непрерывную серию записей.
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
		needWeightReminder = dayStart.Sub(*lastWeightDay) >= weightReminderInterval
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
			"goal_liters": math.Round(float64(wMlGoal)/millilitersPerLiter*waterLitersPrecision) / waterLitersPrecision,
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

// GetAnalytics возвращает историю питания, веса и воды за выбранный период.
func (s *NutritionService) GetAnalytics(ctx context.Context, userID int, days int) (AnalyticsResult, error) {
	if days <= 0 || days > maxAnalyticsDays {
		days = defaultAnalyticsDays
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
