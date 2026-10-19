// Package mlclient предоставляет HTTP-клиент ML-сервисов для генерации персональных тренировочных планов.
package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Plan содержит недельное расписание тренировок, полученное от ML-сервиса.
type Plan struct {
	Split_type string `json:"тип_сплита"`
	Plan_week  []Days `json:"еженедельный_план"`
}

// Days описывает день тренировочного расписания.
type Days struct {
	Day       string     `json:"день"`
	Type_day  string     `json:"тип_дня"`
	Exercises []Muscules `json:"упражнения"`
}

// Muscules описывает мышечную группу и необязательную подгруппу упражнения.
type Muscules struct {
	Group     string  `json:"группа"`
	Sub_group *string `json:"подгруппа"`
}

// Config содержит адреса и таймауты ML-сервисов из переменных окружения с префиксом ML.
type Config struct {
	BaseURL           string        `envconfig:"BASE_URL" required:"true"`
	ChatBaseURL       string        `envconfig:"CHAT_BASE_URL"`
	ChatPath          string        `envconfig:"CHAT_PATH" default:"/ai/user/chat"`
	Timeout           time.Duration `envconfig:"TIMEOUT" default:"30s"`
	ChatTimeout       time.Duration `envconfig:"CHAT_TIMEOUT"`
	GenerationTimeout time.Duration `envconfig:"GENERATION_TIMEOUT" default:"52s"`
}

const (
	defaultPlanTimeout       = 30 * time.Second
	defaultChatTimeout       = 60 * time.Second
	defaultGenerationTimeout = 52 * time.Second
)

// PlanGenerationTimeout возвращает предел времени генерации плана из конфигурации.
func (c Config) PlanGenerationTimeout() time.Duration {
	if c.GenerationTimeout > 0 {
		return c.GenerationTimeout
	}
	return defaultGenerationTimeout
}

// ChatRequestTimeout возвращает таймаут чата с сохранением явно заданного общего таймаута.
func (c Config) ChatRequestTimeout() time.Duration {
	if c.ChatTimeout > 0 {
		return c.ChatTimeout
	}
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultChatTimeout
}

// NewConfig загружает настройки из переменных окружения. Возвращает ошибку при отсутствии обязательного значения или неверном формате.
func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("ML", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

// NewConfigMust загружает настройки из окружения и вызывает панику при ошибке конфигурации.
func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get ML client config: %w", err)
		panic(err)
	}

	return config
}

// ChatURL формирует абсолютный URL ML-чата из настроек адреса и пути.
func (c Config) ChatURL() string {
	base := strings.TrimSpace(c.ChatBaseURL)
	if base == "" {
		base = strings.TrimSpace(c.BaseURL)
	}
	base = strings.TrimSuffix(base, "/")
	path := strings.TrimSpace(c.ChatPath)
	if path == "" {
		path = "/ai/user/chat"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

// Client выполняет HTTP-запросы к ML-сервису.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создаёт клиент с адресом и таймаутом из config.
func NewClient(config Config) *Client {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultPlanTimeout
	}
	return &Client{
		baseURL: strings.TrimSuffix(strings.TrimSpace(config.BaseURL), "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GeneratePlan отправляет reqBody с параметрами пользователя в ML-сервис с контекстом ctx.
// Возвращает план с сопоставленными упражнениями либо ошибку сети, таймаута, HTTP-статуса или декодирования JSON.
func (c *Client) GeneratePlan(ctx context.Context, reqBody any) (*Plan, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("ML base url is not set")
	}

	url := c.baseURL + "/plan/user"
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML returned status %d", resp.StatusCode)
	}

	var plan Plan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	modificatedPlan := replaceExercises(plan)

	return &modificatedPlan, nil
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
