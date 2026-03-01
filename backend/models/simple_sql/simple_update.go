package simplesql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Profile struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	BirthDate       *time.Time `json:"birth_date"`
	Gender          *string    `json:"gender"`
	HeightCm        *int       `json:"height_cm"`
	WeightKg        *int       `json:"weight_kg"`
	ActivityLevel   *string    `json:"activity_level"`
	InjuriesNotes   *string    `json:"injuries_notes"`
	Goal            *string    `json:"goal"`
	FitnessLevel    *string    `json:"fitness_level"`
	TrainingDaysMap []string   `json:"training_days_map"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

func GetProfile(ctx context.Context, conn *pgx.Conn, userID string) (*Profile, error) {
	var p Profile
	err := conn.QueryRow(ctx, `
        SELECT id, user_id, birth_date, gender, height_cm, weight_kg,
               activity_level, injuries_notes, goal, fitness_level,
               training_days_map, created_at, updated_at
        FROM profile
        WHERE user_id = $1
    `, userID).Scan(
		&p.ID, &p.UserID, &p.BirthDate, &p.Gender, &p.HeightCm, &p.WeightKg,
		&p.ActivityLevel, &p.InjuriesNotes, &p.Goal, &p.FitnessLevel,
		&p.TrainingDaysMap, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func UpdateProfile(ctx context.Context, conn *pgx.Conn, userID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	allowed := map[string]bool{
		"birth_date": true, "gender": true, "height_cm": true, "weight_kg": true,
		"activity_level": true, "injuries_notes": true, "goal": true,
		"fitness_level": true, "training_days_map": true,
	}

	updates["updated_at"] = time.Now()

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
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
