package postgres

import "time"

// NutritionEntry represents a food log item consumed by a user.
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

// NutritionFavorite represents a saved favorite food item for quick logging.
type NutritionFavorite struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	ProteinG float64 `json:"protein_g"`
	FatG     float64 `json:"fat_g"`
	CarbsG   float64 `json:"carbs_g"`
	UnitType string  `json:"unit_type"`
}

// NutritionGoal represents daily target macronutrients and calories.
type NutritionGoal struct {
	ProteinG  float64   `json:"protein_g"`
	FatG      float64   `json:"fat_g"`
	CarbsG    float64   `json:"carbs_g"`
	Calories  float64   `json:"calories"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CatalogDish represents a catalog dish item.
type CatalogDish struct {
	ID        int     `json:"id,omitempty"`
	Title     string  `json:"title"`
	ProteinG  float64 `json:"protein_g,omitempty"`
	FatG      float64 `json:"fat_g,omitempty"`
	CarbsG    float64 `json:"carbs_g,omitempty"`
	Calories  float64 `json:"calories,omitempty"`
	BaseGrams float64 `json:"base_grams,omitempty"`
}

// UserProfileData contains fields from sportapp.profile needed to calculate nutritional goals.
type UserProfileData struct {
	Age           *int
	Gender        *string
	HeightCm      *int
	WeightKg      *int
	ActivityLevel *string
	Goal          *string
}

// DailyMacroTotals holds aggregated macronutrient sums.
type DailyMacroTotals struct {
	Protein  float64
	Fat      float64
	Carbs    float64
	Calories float64
}

// FoodDaySummary represents aggregated nutrition intake for a single day.
type FoodDaySummary struct {
	Day      string  `json:"day"`
	ProteinG float64 `json:"protein_g"`
	FatG     float64 `json:"fat_g"`
	CarbsG   float64 `json:"carbs_g"`
	Calories float64 `json:"calories"`
}

// WeightLogSummary represents a weight record on a specific date.
type WeightLogSummary struct {
	Day      string  `json:"day"`
	WeightKg float64 `json:"weight_kg"`
}

// WaterLogSummary represents a water intake record on a specific date.
type WaterLogSummary struct {
	Day      string `json:"day"`
	AmountML int    `json:"amount_ml"`
}
