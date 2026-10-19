package users_postgres_repository

import "time"

// Правила подбора нагрузки и длительности программы.
const (
	programDurationMonths        = 6
	exerciseGroupLimit           = 5
	beginnerWeightMultiplier     = 0.4
	intermediateWeightMultiplier = 0.6
	advancedWeightMultiplier     = 0.8
	highActivityWeightMultiplier = 1.1
	femaleWeightMultiplier       = 0.7
	fullLoadMinimumAge           = 20
	fullLoadMaximumAge           = 50
	ageWeightMultiplier          = 0.8
	injuryWeightMultiplier       = 0.7
	weightLossLoadMultiplier     = 0.9
	weightRoundingPrecision      = 10
	weightComparisonEpsilon      = 1e-9
	percentScale                 = 100
)

// Пороги достижений сохраняют действующие требования к тренировкам.
const (
	fiveWorkoutsThreshold       = 5
	tenWorkoutsThreshold        = 10
	twentyFiveWorkoutsThreshold = 25
	weekWarriorThreshold        = 3
	consistentMonthThreshold    = 12
	achievementWeekWindow       = 7 * 24 * time.Hour
	achievementMonthWindow      = 30 * 24 * time.Hour
	comebackPause               = 14 * 24 * time.Hour
	earlyBirdStartHour          = 5
	earlyBirdEndHour            = 10
	nightOwlStartHour           = 20
	sessionVolume5kKG           = 5000
	sessionVolume10kKG          = 10000
	weightedSetsThreshold       = 20
)

const (
	repBandLow = iota
	repBandNormal
	repBandHigh
)
