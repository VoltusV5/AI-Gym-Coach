package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Plan struct {
	Split_type string `json:"тип_сплита"`
	Plan_week  []Days `json:"еженедельный_план"`
}

type Days struct {
	Day       string     `json:"день"`
	Type_day  string     `json:"тип_дня"`
	Exercises []Muscules `json:"упражнения"`
}

type Muscules struct {
	Group     string  `json:"группа"`
	Sub_group *string `json:"подгруппа"` // null в JSON (например пресс) → nil
}

func GeneratePlan(ctx context.Context, reqBody any) (*Plan, error) {
	url_string := os.Getenv("ML_BASE_URL")
	url := url_string + "/plan/user"
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML returned %d", resp.StatusCode)
	}

	var plan Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, err
	}

	modificated_plan := replaceExercises(plan)

	return &modificated_plan, nil
}

func replaceExercises(plan Plan) Plan {
	newPlan := Plan{
		Split_type: plan.Split_type,
		Plan_week:  make([]Days, len(plan.Plan_week)),
	}

	for i, day := range plan.Plan_week {
		newDay := Days{
			Day:       day.Day,
			Type_day:  day.Type_day,
			Exercises: make([]Muscules, len(day.Exercises)),
		}

		for j, ex := range day.Exercises {
			newEx := Muscules{
				Group:     ex.Group,
				Sub_group: ex.Sub_group,
			}

			if newEx.Sub_group != nil {
				switch *newEx.Sub_group {
				case "Верх спины":
					s := "тяга перед собой широким хватом"
					newEx.Sub_group = &s
				case "Широчайшие спины":
					s := "тяга сверху узким хватом"
					newEx.Sub_group = &s
				}
			}

			newDay.Exercises[j] = newEx
		}

		newPlan.Plan_week[i] = newDay
	}

	return newPlan
}
