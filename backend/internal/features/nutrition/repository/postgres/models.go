package postgres

import "time"

// NutritionEntry описывает запись съеденного продукта в дневнике пользователя.
type NutritionEntry struct {
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

// NutritionFavorite описывает избранный продукт для быстрого добавления в дневник.
type NutritionFavorite struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	ProteinG float64 `json:"protein_g"`
	FatG     float64 `json:"fat_g"`
	CarbsG   float64 `json:"carbs_g"`
	UnitType string  `json:"unit_type"`
}

// NutritionGoal содержит суточные цели по калорийности и БЖУ.
type NutritionGoal struct {
	ProteinG  float64   `json:"protein_g"`
	FatG      float64   `json:"fat_g"`
	CarbsG    float64   `json:"carbs_g"`
	Calories  float64   `json:"calories"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CatalogDish описывает блюдо каталога.
type CatalogDish struct {
	ID        int     `json:"id,omitempty"`
	Title     string  `json:"title"`
	ProteinG  float64 `json:"protein_g,omitempty"`
	FatG      float64 `json:"fat_g,omitempty"`
	CarbsG    float64 `json:"carbs_g,omitempty"`
	Calories  float64 `json:"calories,omitempty"`
	BaseGrams float64 `json:"base_grams,omitempty"`
}

// UserProfileData содержит параметры профиля, необходимые для расчёта целей питания.
type UserProfileData struct {
	Age           *int
	Gender        *string
	HeightCm      *int
	WeightKg      *int
	ActivityLevel *string
	Goal          *string
}

// DailyMacroTotals содержит суммарные значения БЖУ и калорий.
type DailyMacroTotals struct {
	Protein  float64
	Fat      float64
	Carbs    float64
	Calories float64
}

// FoodDaySummary содержит суммарное питание за день.
type FoodDaySummary struct {
	Day      string  `json:"day"`
	ProteinG float64 `json:"protein_g"`
	FatG     float64 `json:"fat_g"`
	CarbsG   float64 `json:"carbs_g"`
	Calories float64 `json:"calories"`
}

// WeightLogSummary содержит запись массы тела на указанную дату.
type WeightLogSummary struct {
	Day      string  `json:"day"`
	WeightKg float64 `json:"weight_kg"`
}

// WaterLogSummary содержит запись потребления воды на указанную дату.
type WaterLogSummary struct {
	Day      string `json:"day"`
	AmountML int    `json:"amount_ml"`
}
