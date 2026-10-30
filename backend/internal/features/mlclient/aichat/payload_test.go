package aichat

import (
	"encoding/json"
	"testing"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonChatIncludesProfileAndPlan(t *testing.T) {
	age, height, weight := 26, 175, 75
	gender, activity, goal, level := "Мужской", "Средняя активность", "Набрать мышцы", "Новичок"
	injuries := false
	profile := users_postgres_repository.Profile{
		Age: &age, HeightCm: &height, WeightKg: &weight, Gender: &gender,
		ActivityLevel: &activity, Goal: &goal, FitnessLevel: &level, InjuriesNotes: &injuries,
		TrainingDaysMap: []string{"Пн", "Ср", "Пт"},
	}
	body, err := buildPythonChatBody("", []ChatMessage{{Role: "user", Content: "Тестовый вопрос"}}, json.RawMessage(`{"тип_сплита":"Фулбади"}`), profile)
	require.NoError(t, err)
	var payload pythonChatRequest
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.JSONEq(t, `{"тип_сплита":"Фулбади"}`, string(payload.Plan))
	assert.Equal(t, map[string]any{
		"возраст": float64(26), "рост": float64(175), "вес": float64(75), "пол": "м",
		"тип_активности": activity, "травмы_или_болезни": "нет", "цель": goal, "уровень_подготовки": "новичок",
		"дни_тренировок": []any{"Пн", "Ср", "Пт"},
	}, payload.User)
	require.Len(t, payload.Messages, 1)
	assert.Equal(t, "Тестовый вопрос", payload.Messages[0].Content)
}
