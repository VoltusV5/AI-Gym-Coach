package users_postgres_repository

import (
	"encoding/json"
	"math"
	"strconv"
)

const (
	highlightWeightEpsilon = 0.05
)

type exerciseSessionMetrics struct {
	maxWeightKg float64
	volumeKg    float64
	minReps     int
}

type SessionHighlight struct {
	ExerciseID   int     `json:"exercise_id"`
	ExerciseName string  `json:"exercise_name"`
	Metric       string  `json:"metric"`
	Previous     float64 `json:"previous"`
	Current      float64 `json:"current"`
	DeltaPercent float64 `json:"delta_percent"`
	MessageKey   string  `json:"message_key"`
	Status       string  `json:"status"`
}

type PlanUpdate struct {
	ExerciseID   int     `json:"exercise_id"`
	ExerciseName string  `json:"exercise_name"`
	OldWeight    float64 `json:"old_weight"`
	NewWeight    float64 `json:"new_weight"`
}

type WorkoutCompleteServiceResult struct {
	NewAchievements   []AchievementUnlocked `json:"new_achievements"`
	SessionHighlights []SessionHighlight    `json:"session_highlights"`
	PlanUpdates       []PlanUpdate          `json:"plan_updates"`
}

func CollectExerciseIDsOrdered(req WorkoutCompleteRequest) []int {
	seen := make(map[int]struct{}, len(req.Slots))
	ids := make([]int, 0, len(req.Slots))
	for _, slot := range req.Slots {
		id := slot.ExerciseID
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func metricsFromSets(sets []SetData) exerciseSessionMetrics {
	var m exerciseSessionMetrics
	m.minReps = 0
	first := true
	for _, s := range sets {
		if s.Reps < 1 {
			continue
		}
		if s.WeightKg > m.maxWeightKg {
			m.maxWeightKg = s.WeightKg
		}
		m.volumeKg += s.WeightKg * float64(s.Reps)
		if first || s.Reps < m.minReps {
			m.minReps = s.Reps
			first = false
		}
	}
	return m
}

func mergeExerciseMetrics(a, b exerciseSessionMetrics) exerciseSessionMetrics {
	out := a
	if b.maxWeightKg > out.maxWeightKg {
		out.maxWeightKg = b.maxWeightKg
	}
	out.volumeKg += b.volumeKg
	if b.minReps < out.minReps || out.minReps == 0 {
		out.minReps = b.minReps
	}
	return out
}

func aggregateCurrentByExercise(req WorkoutCompleteRequest) map[int]exerciseSessionMetrics {
	m := make(map[int]exerciseSessionMetrics)
	for _, slot := range req.Slots {
		cur := metricsFromSets(slot.Sets)
		if cur.minReps == 0 && cur.maxWeightKg == 0 && cur.volumeKg == 0 {
			continue
		}
		id := slot.ExerciseID
		if prev, ok := m[id]; ok {
			m[id] = mergeExerciseMetrics(prev, cur)
		} else {
			m[id] = cur
		}
	}
	return m
}

func repIntensityBand(minReps int) int {
	switch {
	case minReps < normRepsMin:
		return 0
	case minReps > normRepsMax:
		return 2
	default:
		return 1
	}
}

func findPreviousExerciseMetrics(
	history []WorkoutCompleteRequest,
	exerciseID int,
) *exerciseSessionMetrics {
	for i := len(history) - 1; i >= 0; i-- {
		w := history[i]
		var merged *exerciseSessionMetrics
		for _, slot := range w.Slots {
			if slot.ExerciseID != exerciseID {
				continue
			}
			cur := metricsFromSets(slot.Sets)
			if cur.minReps == 0 && cur.maxWeightKg == 0 && cur.volumeKg == 0 {
				continue
			}
			if merged == nil {
				m := cur
				merged = &m
			} else {
				next := mergeExerciseMetrics(*merged, cur)
				merged = &next
			}
		}
		if merged != nil {
			return merged
		}
	}
	return nil
}

func roundHighlightFloat(x float64) float64 {
	return math.Round(x*10) / 10
}

func BuildSessionHighlights(
	history []WorkoutCompleteRequest,
	current WorkoutCompleteRequest,
	names map[int]string,
) []SessionHighlight {
	curByID := aggregateCurrentByExercise(current)
	if len(curByID) == 0 {
		return nil
	}

	var idsOrdered []int
	seen := make(map[int]struct{}, len(curByID))
	for _, slot := range current.Slots {
		id := slot.ExerciseID
		if _, ok := curByID[id]; !ok {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		idsOrdered = append(idsOrdered, id)
	}

	var out []SessionHighlight
	for _, exerciseID := range idsOrdered {
		cur := curByID[exerciseID]
		prev := findPreviousExerciseMetrics(history, exerciseID)
		name := names[exerciseID]

		if prev == nil {
			out = append(out, SessionHighlight{
				ExerciseID:   exerciseID,
				ExerciseName: name,
				Metric:       "max_weight_kg",
				Previous:     0,
				Current:      roundHighlightFloat(cur.maxWeightKg),
				DeltaPercent: 0,
				MessageKey:   "new_record",
				Status:       "record",
			})
			continue
		}

		weightUp := cur.maxWeightKg > prev.maxWeightKg+1e-9 && prev.maxWeightKg > 0
		weightDown := cur.maxWeightKg < prev.maxWeightKg-highlightWeightEpsilon

		if weightUp {
			pct := (cur.maxWeightKg - prev.maxWeightKg) / prev.maxWeightKg * 100
			bCur := repIntensityBand(cur.minReps)
			bPrev := repIntensityBand(prev.minReps)
			msg := "weight_pr"
			status := "pr"
			if bCur == bPrev {
				msg = "weight_up_percent"
				status = "increase"
			}
			out = append(out, SessionHighlight{
				ExerciseID:   exerciseID,
				ExerciseName: name,
				Metric:       "max_weight_kg",
				Previous:     roundHighlightFloat(prev.maxWeightKg),
				Current:      roundHighlightFloat(cur.maxWeightKg),
				DeltaPercent: roundHighlightFloat(pct),
				MessageKey:   msg,
				Status:       status,
			})
			continue
		}

		if weightDown {
			out = append(out, SessionHighlight{
				ExerciseID:   exerciseID,
				ExerciseName: name,
				Metric:       "max_weight_kg",
				Previous:     roundHighlightFloat(prev.maxWeightKg),
				Current:      roundHighlightFloat(cur.maxWeightKg),
				DeltaPercent: 0,
				MessageKey:   "weight_maintained",
				Status:       "record",
			})
			continue
		}
		sameWeight := math.Abs(cur.maxWeightKg-prev.maxWeightKg) <= highlightWeightEpsilon
		if sameWeight {
			if cur.volumeKg > prev.volumeKg+1e-9 && prev.volumeKg > 0 {
				pct := (cur.volumeKg - prev.volumeKg) / prev.volumeKg * 100
				out = append(out, SessionHighlight{
					ExerciseID:   exerciseID,
					ExerciseName: name,
					Metric:       "volume_kg",
					Previous:     roundHighlightFloat(prev.volumeKg),
					Current:      roundHighlightFloat(cur.volumeKg),
					DeltaPercent: roundHighlightFloat(pct),
					MessageKey:   "volume_up_percent",
					Status:       "increase",
				})
			} else {
				out = append(out, SessionHighlight{
					ExerciseID:   exerciseID,
					ExerciseName: name,
					Metric:       "max_weight_kg",
					Previous:     roundHighlightFloat(prev.maxWeightKg),
					Current:      roundHighlightFloat(cur.maxWeightKg),
					DeltaPercent: 0,
					MessageKey:   "weight_maintained",
					Status:       "record",
				})
			}
		}
	}

	return out
}

func BuildPlanUpdates(
	oldWeights map[string]float64,
	newWeightsJSON []byte,
	names map[int]string,
) []PlanUpdate {
	var newWeights map[string]float64
	if err := json.Unmarshal(newWeightsJSON, &newWeights); err != nil {
		return nil
	}

	var out []PlanUpdate
	for k, newVal := range newWeights {
		oldVal := oldWeights[k]
		if math.Abs(newVal-oldVal) < 1e-9 {
			continue
		}
		id, _ := strconv.Atoi(k)
		out = append(out, PlanUpdate{
			ExerciseID:   id,
			ExerciseName: names[id],
			OldWeight:    roundHighlightFloat(oldVal),
			NewWeight:    roundHighlightFloat(newVal),
		})
	}
	return out
}