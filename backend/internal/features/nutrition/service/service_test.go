package service

import (
	"math"
	"testing"
)

func TestCaloriesFromMacros(t *testing.T) {
	tests := []struct {
		name     string
		p, f, c  float64
		expected float64
	}{
		{"zeros", 0, 0, 0, 0},
		{"standard", 10, 10, 10, 170}, // 10*4 + 10*9 + 10*4 = 170
		{"protein only", 25, 0, 0, 100},
		{"fat only", 0, 10, 0, 90},
		{"carbs only", 0, 0, 50, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CaloriesFromMacros(tt.p, tt.f, tt.c)
			if math.Abs(got-tt.expected) > 1e-6 {
				t.Errorf("CaloriesFromMacros(%v, %v, %v) = %v; want %v", tt.p, tt.f, tt.c, got, tt.expected)
			}
		})
	}
}

func TestScaleFrom100(t *testing.T) {
	tests := []struct {
		name     string
		per100   float64
		grams    float64
		expected float64
	}{
		{"standard 200g", 20, 200, 40},
		{"50g portion", 15, 50, 7.5},
		{"zero grams defaults to 100g", 25, 0, 25},
		{"negative grams defaults to 100g", 30, -10, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScaleFrom100(tt.per100, tt.grams)
			if math.Abs(got-tt.expected) > 1e-6 {
				t.Errorf("ScaleFrom100(%v, %v) = %v; want %v", tt.per100, tt.grams, got, tt.expected)
			}
		})
	}
}

func TestDailyWaterGoalMlFromWeight(t *testing.T) {
	tests := []struct {
		name     string
		weightKg float64
		expected int
	}{
		{"default for zero weight", 0, 2000},
		{"default for negative weight", -10, 2000},
		{"minimum clamp (30kg * 33 = 990 -> min 1500)", 30, 1500},
		{"normal weight 70kg (70 * 33 = 2310)", 70, 2310},
		{"maximum clamp (150kg * 33 = 4950 -> max 4500)", 150, 4500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DailyWaterGoalMlFromWeight(tt.weightKg)
			if got != tt.expected {
				t.Errorf("DailyWaterGoalMlFromWeight(%v) = %d; want %d", tt.weightKg, got, tt.expected)
			}
		})
	}
}

func TestMealTypeOrDefault(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Breakfast", "breakfast"},
		{"LUNCH", "lunch"},
		{"dinner", "dinner"},
		{"Snack", "snack"},
		{"", "snack"},
		{"invalid", "snack"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := MealTypeOrDefault(tt.input); got != tt.expected {
				t.Errorf("MealTypeOrDefault(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCalculateGoalFromProfile(t *testing.T) {
	// 30 yo, 175cm, 70kg, male, light activity, maintain
	goal := CalculateGoalFromProfile(30, 175, 70, "Мужчина", "лёгкая активность (физические нагрузки 1-3 раза в неделю)", "Поддержание")

	if goal.Calories < 1200 {
		t.Errorf("Goal calories = %v; expected >= 1200", goal.Calories)
	}
	if goal.ProteinG <= 0 {
		t.Errorf("Goal protein = %v; expected > 0", goal.ProteinG)
	}
	if goal.FatG <= 0 {
		t.Errorf("Goal fat = %v; expected > 0", goal.FatG)
	}
	if goal.CarbsG < 0 {
		t.Errorf("Goal carbs = %v; expected >= 0", goal.CarbsG)
	}

	// Female profile
	femaleGoal := CalculateGoalFromProfile(25, 165, 55, "Женщина", "средняя активность (физические нагрузки 3-5 раза в неделю)", "Сбросить вес")
	if femaleGoal.Calories < 1200 {
		t.Errorf("Female calories should be at least 1200, got %v", femaleGoal.Calories)
	}
	if femaleGoal.ProteinG != 55*1.2 {
		t.Errorf("Weight loss protein expected %v, got %v", 55*1.2, femaleGoal.ProteinG)
	}
}

func TestDayBoundsFromQuery(t *testing.T) {
	svc := NewNutritionService(nil)

	start, end, dayStr := svc.DayBoundsFromQuery("2026-10-15")
	if dayStr != "2026-10-15" {
		t.Errorf("Expected dayStr 2026-10-15, got %s", dayStr)
	}
	if end.Sub(start).Hours() != 24 {
		t.Errorf("Expected 24 hour interval, got %v", end.Sub(start).Hours())
	}
}
