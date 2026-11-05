package workouts_domain

import (
	"fmt"
	"math"
	"sort"
	"time"

	"sport_app/internal/core/transport/http/v2request"
)

func Validate(input Input, now time.Time) (Input, error) {
	var err error
	input.ClientSessionID, err = v2request.UUID(input.ClientSessionID)
	if err != nil {
		return input, err
	}
	input.PlanID, err = v2request.UUID(input.PlanID)
	if err != nil {
		return input, err
	}
	input.PlanDayID, err = v2request.UUID(input.PlanDayID)
	if err != nil {
		return input, err
	}
	if input.PlanVersion < 1 || input.StartedAt.IsZero() || input.FinishedAt.Before(input.StartedAt) || input.FinishedAt.After(now.Add(time.Minute)) || input.StartedAt.Year() < 1 || input.FinishedAt.Year() > 9999 || len(input.Slots) == 0 {
		return input, v2request.Invalid()
	}
	positions := map[int]bool{}
	for _, slot := range input.Slots {
		if slot.ExerciseID < 1 || slot.ExerciseID > math.MaxInt32 || slot.Position < 0 || slot.Position > math.MaxInt32 || positions[slot.Position] || len(slot.Sets) == 0 {
			return input, v2request.Invalid()
		}
		positions[slot.Position] = true
		for _, set := range slot.Sets {
			if set.Reps < 1 || set.Reps > math.MaxInt32 || set.WeightKg < 0 || math.IsNaN(set.WeightKg) || math.IsInf(set.WeightKg, 0) || math.IsInf(set.WeightKg*float64(set.Reps), 0) {
				return input, v2request.Invalid()
			}
		}
	}
	if math.IsInf(Measure(input).VolumeKg, 0) {
		return input, v2request.Invalid()
	}
	input.StartedAt = input.StartedAt.UTC()
	input.FinishedAt = input.FinishedAt.UTC()
	return input, nil
}
func Measure(input Input) Metrics {
	out := Metrics{DurationSec: int64(input.FinishedAt.Sub(input.StartedAt) / time.Second)}
	exercises := map[int]bool{}
	for _, slot := range input.Slots {
		for _, set := range slot.Sets {
			if !set.IsWarmup {
				out.VolumeKg += set.WeightKg * float64(set.Reps)
				out.WorkingSets++
				exercises[slot.ExerciseID] = true
			}
		}
	}
	out.ExerciseCount = len(exercises)
	return out
}
func ExercisePoint(w Workout, id int) (Point, bool) {
	p := Point{SessionID: w.ID, FinishedAt: w.FinishedAt}
	found := false
	for _, slot := range w.Slots {
		if slot.ExerciseID != id {
			continue
		}
		for _, set := range slot.Sets {
			if set.IsWarmup {
				continue
			}
			found = true
			p.VolumeKg += set.WeightKg * float64(set.Reps)
			if set.WeightKg > p.WorkingWeightKg || (set.WeightKg == p.WorkingWeightKg && set.Reps > p.Reps) {
				p.WorkingWeightKg = set.WeightKg
				p.Reps = set.Reps
			}
		}
	}
	return p, found
}
func Records(current Workout, previous []Workout) []Highlight {
	highlights := []Highlight{}
	ids := map[int]bool{}
	ordered := []int{}
	for _, slot := range current.Slots {
		if !ids[slot.ExerciseID] {
			ordered = append(ordered, slot.ExerciseID)
		}
		ids[slot.ExerciseID] = true
	}
	sort.Ints(ordered)
	for _, id := range ordered {
		p, found := ExercisePoint(current, id)
		if !found {
			continue
		}
		weight, volume := 0.0, 0.0
		hasPrevious := false
		for _, w := range previous {
			if !w.FinishedAt.Before(current.FinishedAt) {
				continue
			}
			q, ok := ExercisePoint(w, id)
			if ok {
				hasPrevious = true
				weight = math.Max(weight, q.WorkingWeightKg)
				volume = math.Max(volume, q.VolumeKg)
			}
		}
		if !hasPrevious {
			continue
		}
		for _, m := range []struct {
			name       string
			old, value float64
		}{{"working_weight_kg", weight, p.WorkingWeightKg}, {"volume_kg", volume, p.VolumeKg}} {
			if m.value <= m.old {
				continue
			}
			old := m.old
			h := Highlight{ExerciseID: &id, Metric: m.name, Previous: &old, Current: m.value, Message: fmt.Sprintf("New exercise record: %s", m.name)}
			if old > 0 {
				delta := 100 * (m.value - old) / old
				h.DeltaPercent = &delta
			}
			highlights = append(highlights, h)
		}
	}
	return highlights
}
func ExerciseHistory(items []Workout, id int) ExerciseProgress {
	out := ExerciseProgress{ExerciseID: id, Points: []Point{}, Records: []Highlight{}}
	sort.SliceStable(items, func(i, j int) bool { return items[i].FinishedAt.Before(items[j].FinishedAt) })
	previous := []Workout{}
	for _, w := range items {
		if p, ok := ExercisePoint(w, id); ok {
			out.Points = append(out.Points, p)
		}
		for _, h := range Records(w, previous) {
			if h.ExerciseID != nil && *h.ExerciseID == id {
				out.Records = append(out.Records, h)
			}
		}
		previous = append(previous, w)
	}
	return out
}
func Bounds(period string, anchor time.Time, loc *time.Location) (time.Time, time.Time, error) {
	anchor = anchor.In(loc)
	start := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, loc)
	switch period {
	case "week":
		start = start.AddDate(0, 0, -(int(start.Weekday())+6)%7)
		return start, start.AddDate(0, 0, 7), nil
	case "month":
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 1, 0), nil
	case "all":
		return time.Date(1, 1, 1, 0, 0, 0, 0, loc), start.AddDate(0, 0, 1), nil
	default:
		return time.Time{}, time.Time{}, v2request.Invalid()
	}
}
func dateStart(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}
func scheduleCounts(data Data, start, end, now time.Time, _ *time.Location) (int, int) {
	if end.After(now) {
		end = now
	}
	groups := map[string][]Revision{}
	for _, revision := range data.Revisions {
		groups[revision.Plan.ID] = append(groups[revision.Plan.ID], revision)
	}
	planned, completed := 0, 0
	for _, revisions := range groups {
		sort.SliceStable(revisions, func(i, j int) bool { return revisions[i].At.Before(revisions[j].At) })
		zones := map[string]bool{}
		for _, r := range revisions {
			zones[r.Plan.Schedule.TimeZone] = true
		}
		counted := map[string]bool{}
		for zone := range zones {
			loc, err := time.LoadLocation(zone)
			if err != nil {
				continue
			}
			first := dateStart(revisions[0].At, loc)
			begin := dateStart(start, loc)
			if begin.Before(first) {
				begin = first
			}
			for day := begin; day.Before(end); day = day.AddDate(0, 0, 1) {
				next := day.AddDate(0, 0, 1)
				var state *Revision
				for i := range revisions {
					if revisions[i].At.After(day) {
						break
					}
					state = &revisions[i]
				}
				occurrence := day
				if state == nil || !state.Plan.IsActive {
					state = nil
					for i := range revisions {
						r := &revisions[i]
						if !r.At.Before(day) && r.At.Before(next) && r.At.Before(end) && r.Plan.IsActive {
							state = r
							occurrence = r.At
							break
						}
					}
				}
				if state == nil || !state.Plan.IsActive || state.Plan.Schedule.TimeZone != zone || occurrence.Before(start) || !occurrence.Before(end) {
					continue
				}
				date := day.Format(time.DateOnly)
				if date < state.Plan.Schedule.StartDate {
					continue
				}
				weekday := (int(day.Weekday())+6)%7 + 1
				for _, entry := range state.Plan.Schedule.Entries {
					key := date + ":" + entry.PlanDayID
					if entry.Weekday != weekday || counted[key] {
						continue
					}
					counted[key] = true
					planned++
					for _, w := range data.Workouts {
						if w.PlanID == state.Plan.ID && w.PlanDayID == entry.PlanDayID && w.FinishedAt.In(loc).Format(time.DateOnly) == date && !w.FinishedAt.Before(start) && w.FinishedAt.Before(end) {
							completed++
							break
						}
					}
				}
			}
		}
	}
	return planned, completed
}

func Aggregate(data Data, period string, start, end, now time.Time, loc *time.Location) Progress {
	out := Progress{Period: period, Start: start.UTC(), End: end.UTC(), TimeZone: loc.String(), Highlights: []Highlight{}, ComputedAt: now.UTC()}
	if period == "all" {
		first := now
		for _, r := range data.Revisions {
			if r.Plan.IsActive && r.At.Before(first) {
				first = r.At
			}
		}
		for _, w := range data.Workouts {
			if w.FinishedAt.Before(first) {
				first = w.FinishedAt
			}
		}
		for _, w := range data.Weights {
			d, err := time.ParseInLocation(time.DateOnly, w.Day, loc)
			if err == nil && d.Before(first) {
				first = d
			}
		}
		start = dateStart(first, loc)
		out.Start = start.UTC()
	}
	weights := []Weight{}
	for _, w := range data.Weights {
		d, err := time.ParseInLocation(time.DateOnly, w.Day, loc)
		if err == nil && !d.Before(start) && d.Before(end) {
			weights = append(weights, w)
		}
	}
	sort.Slice(weights, func(i, j int) bool { return weights[i].Day < weights[j].Day })
	if len(weights) > 1 {
		delta := weights[len(weights)-1].WeightKg - weights[0].WeightKg
		out.WeightChangeKg = &delta
	}
	previous := []Workout{}
	items := append([]Workout(nil), data.Workouts...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].FinishedAt.Before(items[j].FinishedAt) })
	for _, w := range items {
		if !w.FinishedAt.Before(start) && w.FinishedAt.Before(end) {
			out.WorkoutCount++
			out.VolumeKg += Measure(w.Input).VolumeKg
			out.Highlights = append(out.Highlights, Records(w, previous)...)
		}
		previous = append(previous, w)
	}
	planned, completed := scheduleCounts(data, start, end, now, loc)
	if planned > 0 {
		adherence := float64(completed) / float64(planned)
		out.Adherence = &adherence
	}
	cutoff := end
	if cutoff.After(now) {
		cutoff = now
	}
	week, _, _ := Bounds("week", cutoff, loc)
	for len(data.Revisions) > 0 {
		previousWeek := week.AddDate(0, 0, -7)
		p, c := scheduleCounts(data, previousWeek, week, now, loc)
		if p == 0 || c < p {
			break
		}
		out.StreakWeeks++
		week = previousWeek
	}
	return out
}
