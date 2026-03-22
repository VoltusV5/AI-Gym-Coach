package simplesql

import (
	"context"
	"fmt"
	"log"
	"math"
	"sport_app/mlclient"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Profile struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	Age             int        `json:"age"`
	Gender          *string    `json:"gender"`
	HeightCm        *int       `json:"height_cm"`
	WeightKg        *int       `json:"weight_kg"`
	ActivityLevel   *string    `json:"activity_level"`
	InjuriesNotes   bool       `json:"injuries_notes"`
	Goal            *string    `json:"goal"`
	FitnessLevel    *string    `json:"fitness_level"`
	TrainingDaysMap []string   `json:"training_days_map"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type ExWithWeight struct {
	ID     int    `json:"id"`
	EXName string `json:"exercise_name"`
	Weight *int   `json:"weight"`
}

type EDaysWithWeight struct {
	Day       string           `json:"day"`
	DName     string           `json:"day_name"`
	Exercises [][]ExWithWeight `json:"exercises"`
}

type EPlanWithWeight struct {
	Split string            `json:"split"`
	Plan  []EDaysWithWeight `json:"plan"`
}

type ExNoWeight struct {
	ID     int    `json:"id"`
	EXName string `json:"exercise_name"`
}

type EDaysNoWeight struct {
	Day       string         `json:"day"`
	DName     string         `json:"day_name"`
	Exercises [][]ExNoWeight `json:"exercises"`
}

type EPlanNoWeight struct {
	Split string          `json:"split"`
	Plan  []EDaysNoWeight `json:"plan"`
}

func GetProfile(ctx context.Context, conn *pgxpool.Pool, userID string) (*Profile, error) {
	var p Profile
	err := conn.QueryRow(ctx, `
        SELECT id, user_id, age, gender, height_cm, weight_kg,
               activity_level, injuries_notes, goal, fitness_level,
               training_days_map, created_at, updated_at
        FROM profile
        WHERE user_id = $1
    `, userID).Scan(
		&p.ID, &p.UserID, &p.Age, &p.Gender, &p.HeightCm, &p.WeightKg,
		&p.ActivityLevel, &p.InjuriesNotes, &p.Goal, &p.FitnessLevel,
		&p.TrainingDaysMap, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func UpdateProfile(ctx context.Context, conn *pgxpool.Pool, userID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	allowed := map[string]bool{
		"age": true, "gender": true, "height_cm": true, "weight_kg": true,
		"activity_level": true, "injuries_notes": true, "goal": true,
		"fitness_level": true, "training_days_map": true,
	}

	updates["updated_at"] = time.Now()

	setParts := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)+1)
	idx := 1

	for field, value := range updates {
		if !allowed[field] && field != "updated_at" {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, idx))
		args = append(args, value)
		idx++
	}

	if len(setParts) == 0 {
		return nil
	}

	args = append(args, userID)
	query := fmt.Sprintf(`
        UPDATE profile
        SET %s
        WHERE user_id = $%d
    `, strings.Join(setParts, ", "), idx)

	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("User profile %s not found", userID)
	}
	return nil
}

func GetExercises(ctx context.Context, conn *pgxpool.Pool, plan mlclient.Plan, userID string) (EPlanWithWeight, EPlanNoWeight) {
	eplanWithWeight := EPlanWithWeight{
		Split: plan.Split_type,
		Plan:  make([]EDaysWithWeight, len(plan.Plan_week)),
	}
	eplanNoWeight := EPlanNoWeight{
		Split: plan.Split_type,
		Plan:  make([]EDaysNoWeight, len(plan.Plan_week)),
	}

	for i, day := range plan.Plan_week {
		edayWithWeight := EDaysWithWeight{
			Day:       day.Day,
			DName:     day.Type_day,
			Exercises: make([][]ExWithWeight, len(day.Exercises)),
		}
		edayNoWeight := EDaysNoWeight{
			Day:       day.Day,
			DName:     day.Type_day,
			Exercises: make([][]ExNoWeight, len(day.Exercises)),
		}

		for j, ex := range day.Exercises {
			rows, err := conn.Query(ctx, `
				SELECT id, exercises_name, working_weights
				FROM exercises
				WHERE muscular_group = $1 AND muscular_subgroup = $2
				ORDER BY id
				LIMIT 5
			`, ex.Group, ex.Sub_group)
			if err != nil {
				log.Printf("Error request for  %s/%s: %v", ex.Group, ex.Sub_group, err)
				edayWithWeight.Exercises[j] = []ExWithWeight{}
				edayNoWeight.Exercises[j] = []ExNoWeight{}
				continue
			}

			var withWeight []ExWithWeight
			var noWeight []ExNoWeight
			for rows.Next() {
				var id int
				var name string
				var weight *int
				if err := rows.Scan(&id, &name, &weight); err != nil {
					log.Printf("Error scaning: %v", err)
					continue
				}

				if weight != nil {
					countWeight(ctx, conn, weight, userID)
				}

				withWeight = append(withWeight, ExWithWeight{ID: id, EXName: name, Weight: weight})
				noWeight = append(noWeight, ExNoWeight{ID: id, EXName: name})
			}
			rows.Close()
			edayWithWeight.Exercises[j] = withWeight
			edayNoWeight.Exercises[j] = noWeight
		}
		eplanWithWeight.Plan[i] = edayWithWeight
		eplanNoWeight.Plan[i] = edayNoWeight
	}

	return eplanWithWeight, eplanNoWeight
}

func countWeight(ctx context.Context, conn *pgxpool.Pool, weight *int, userID string) {
	profile, err := GetProfile(ctx, conn, userID)
	if err != nil {
		panic(err)
	}

	k := 1.0
	if profile.FitnessLevel == String("Новичок") {
		k *= 0.4
	} else if profile.FitnessLevel == String("Любитель") {
		k *= 0.6
	} else {
		k *= 0.8
	}

	if profile.ActivityLevel == String("Высокая") || profile.ActivityLevel == String("Очень высокая") {
		k *= 1.1
	}

	if profile.Gender == String("Женщина") {
		k *= 0.7
	}

	if profile.Age < 20 || profile.Age > 50 {
		k *= 0.8
	}

	if profile.InjuriesNotes == true {
		k *= 0.7
	}

	if profile.Goal == String("Сжечь жир") || profile.Goal == String("Сбросить вес") {
		k *= 0.9
	}

	*weight = int(math.Floor(float64(*weight) * k))
}
