package service

import "time"

// Коэффициенты действующих формул питания; выделение констант не меняет расчёты.
const (
	proteinCaloriesPerGram      = 4
	fatCaloriesPerGram          = 9
	carbCaloriesPerGram         = 4
	referencePortionGrams       = 100
	sedentaryActivityMultiplier = 1.2
	lightActivityMultiplier     = 1.375
	moderateActivityMultiplier  = 1.55
	highActivityMultiplier      = 1.725
	veryHighActivityMultiplier  = 1.9
	targetCaloriesPerKg         = 2
	targetDeltaCaloriesPerKg    = 220
	bmrWeightCoefficient        = 10
	bmrHeightCoefficient        = 6.25
	bmrAgeCoefficient           = 5
	bmrFemaleOffset             = 161
	bmrMaleOffset               = 5
	minimumDailyCalories        = 1200
	weightLossProteinPerKg      = 1.2
	muscleGainProteinPerKg      = 1.8
	maintenanceProteinPerKg     = 1.0
	weightLossFatFraction       = 0.2
	muscleGainFatFraction       = 0.25
	maintenanceFatFraction      = 0.3
	recalculatedFatFraction     = 0.28
	defaultWaterGoalML          = 2000
	minWaterGoalML              = 1500
	maxWaterGoalML              = 4500
	waterMLPerKg                = 33
	millilitersPerLiter         = 1000
	waterLitersPrecision        = 100
	defaultProfileAge           = 30
	maximumProfileAge           = 120
	defaultProfileHeightCM      = 175
	minimumProfileHeightCM      = 80
	maximumProfileHeightCM      = 250
	defaultProfileWeightKG      = 70
	minimumProfileWeightKG      = 30
	maximumProfileWeightKG      = 400
	targetDeltaEpsilon          = 1e-6
	nutritionDayDuration        = 24 * time.Hour
	defaultMealHour             = 12
	statsWeekDays               = 7
	statsMonthDays              = 30
	weightReminderInterval      = 72 * time.Hour
	maxDishSearchLimit          = 200
	maxAnalyticsDays            = 4000
	defaultAnalyticsDays        = 30
)

// DefaultDishSearchLimit задаёт размер поиска по каталогу без явного лимита.
const DefaultDishSearchLimit = 30
