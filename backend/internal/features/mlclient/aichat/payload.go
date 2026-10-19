package aichat

import (
	"encoding/json"
	"strings"

	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
)

type pythonChatRequest struct {
	Plan     json.RawMessage `json:"plan"`
	User     map[string]any  `json:"user"`
	Model    string          `json:"model,omitempty"`
	Messages []ChatMessage   `json:"messages"`
}

func buildPythonChatBody(
	model string,
	messages []ChatMessage,
	planJSON json.RawMessage,
	profile users_postgres_repository.Profile,
) ([]byte, error) {
	if len(planJSON) == 0 {
		planJSON = json.RawMessage(`{}`)
	}
	body := pythonChatRequest{
		Plan:     planJSON,
		User:     profileToPythonUser(profile),
		Model:    strings.TrimSpace(model),
		Messages: normalizeMessages(messages),
	}
	return json.Marshal(body)
}

func profileToPythonUser(p users_postgres_repository.Profile) map[string]any {
	m := make(map[string]any)
	if p.Age != nil {
		m["возраст"] = *p.Age
	}
	if p.HeightCm != nil {
		m["рост"] = *p.HeightCm
	}
	if p.WeightKg != nil {
		m["вес"] = *p.WeightKg
	}
	if p.Gender != nil && strings.TrimSpace(*p.Gender) != "" {
		m["пол"] = genderShort(*p.Gender)
	}
	if p.ActivityLevel != nil && strings.TrimSpace(*p.ActivityLevel) != "" {
		m["тип_активности"] = strings.TrimSpace(*p.ActivityLevel)
	}
	m["травмы_или_болезни"] = injuriesText(p.InjuriesNotes)
	if p.Goal != nil && strings.TrimSpace(*p.Goal) != "" {
		m["цель"] = strings.TrimSpace(*p.Goal)
	}
	if p.FitnessLevel != nil && strings.TrimSpace(*p.FitnessLevel) != "" {
		m["уровень_подготовки"] = strings.ToLower(strings.TrimSpace(*p.FitnessLevel))
	}
	if len(p.TrainingDaysMap) > 0 {
		m["дни_тренировок"] = append([]string(nil), p.TrainingDaysMap...)
	}
	return m
}

func genderShort(g string) string {
	s := strings.ToLower(strings.TrimSpace(g))
	if strings.HasPrefix(s, "ж") {
		return "ж"
	}
	if strings.HasPrefix(s, "м") {
		return "м"
	}
	return strings.TrimSpace(g)
}

func injuriesText(v *bool) string {
	if v == nil {
		return "не указано"
	}
	if *v {
		return "да"
	}
	return "нет"
}
