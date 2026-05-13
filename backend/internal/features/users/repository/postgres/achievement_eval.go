package users_postgres_repository

import (
	"encoding/json"
	"sort"
	"strings"
	"time"
)

const (
	achFirstWorkout       = "first_workout"
	achFiveWorkouts       = "five_workouts"
	achTenWorkouts        = "ten_workouts"
	achTwentyFiveWorkouts = "twenty_five_workouts"
	achWeekWarrior        = "week_warrior"
	achConsistentMonth    = "consistent_month"
	achComeback           = "comeback"
	achEarlyBird          = "early_bird"
	achNightOwl           = "night_owl"
	achVolume5k           = "volume_session_5k"
	achVolume10k          = "volume_session_10k"
	achDoubleDigitSets    = "double_digit_sets"
	achProfileReady       = "profile_ready"
)

func parseCompletedWorkoutsJSON(raw []byte) ([]WorkoutCompleteRequest, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	out := make([]WorkoutCompleteRequest, 0, len(items))
	for _, it := range items {
		var w WorkoutCompleteRequest
		if err := json.Unmarshal(it, &w); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func workoutVolumeKg(w WorkoutCompleteRequest) float64 {
	var s float64
	for _, slot := range w.Slots {
		for _, set := range slot.Sets {
			s += set.WeightKg * float64(set.Reps)
		}
	}
	return s
}

func weightedSetsInSession(w WorkoutCompleteRequest) int {
	n := 0
	for _, slot := range w.Slots {
		for _, set := range slot.Sets {
			if set.WeightKg > 0 && set.Reps > 0 {
				n++
			}
		}
	}
	return n
}

func profileReadyForAchievements(p Profile) bool {
	if p.Age == nil {
		return false
	}
	if p.Gender == nil || strings.TrimSpace(*p.Gender) == "" {
		return false
	}
	if p.HeightCm == nil || p.WeightKg == nil {
		return false
	}
	if p.ActivityLevel == nil || strings.TrimSpace(*p.ActivityLevel) == "" {
		return false
	}
	if p.InjuriesNotes == nil {
		return false
	}
	if p.Goal == nil || strings.TrimSpace(*p.Goal) == "" {
		return false
	}
	if p.FitnessLevel == nil || strings.TrimSpace(*p.FitnessLevel) == "" {
		return false
	}
	if len(p.TrainingDaysMap) == 0 {
		return false
	}
	return true
}

func earnedAchievementTitles(
	workouts []WorkoutCompleteRequest,
	profile Profile,
	now time.Time,
) map[string]struct{} {
	earned := map[string]struct{}{}
	n := len(workouts)
	if n == 0 {
		if profileReadyForAchievements(profile) {
			earned[achProfileReady] = struct{}{}
		}
		return earned
	}

	if n >= 1 {
		earned[achFirstWorkout] = struct{}{}
	}
	if n >= 5 {
		earned[achFiveWorkouts] = struct{}{}
	}
	if n >= 10 {
		earned[achTenWorkouts] = struct{}{}
	}
	if n >= 25 {
		earned[achTwentyFiveWorkouts] = struct{}{}
	}

	cutoff7 := now.Add(-7 * 24 * time.Hour)
	c7 := 0
	cutoff30 := now.Add(-30 * 24 * time.Hour)
	c30 := 0
	for _, w := range workouts {
		if !w.FinishedAt.Before(cutoff7) {
			c7++
		}
		if !w.FinishedAt.Before(cutoff30) {
			c30++
		}
	}
	if c7 >= 3 {
		earned[achWeekWarrior] = struct{}{}
	}
	if c30 >= 12 {
		earned[achConsistentMonth] = struct{}{}
	}

	sorted := append([]WorkoutCompleteRequest(nil), workouts...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].FinishedAt.Before(sorted[j].FinishedAt)
	})
	pause := 14 * 24 * time.Hour
	for i := 1; i < len(sorted); i++ {
		if sorted[i].FinishedAt.Sub(sorted[i-1].FinishedAt) >= pause {
			earned[achComeback] = struct{}{}
			break
		}
	}

	for _, w := range workouts {
		h := w.FinishedAt.Hour()
		if h >= 5 && h < 10 {
			earned[achEarlyBird] = struct{}{}
		}
		if h >= 20 || h < 5 {
			earned[achNightOwl] = struct{}{}
		}
		v := workoutVolumeKg(w)
		if v >= 5000 {
			earned[achVolume5k] = struct{}{}
		}
		if v >= 10000 {
			earned[achVolume10k] = struct{}{}
		}
		if weightedSetsInSession(w) >= 20 {
			earned[achDoubleDigitSets] = struct{}{}
		}
	}

	if profileReadyForAchievements(profile) {
		earned[achProfileReady] = struct{}{}
	}

	return earned
}