package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kelseyhightower/envconfig"
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
	Sub_group *string `json:"подгруппа"`
}

type Config struct {
	BaseURL     string `envconfig:"BASE_URL" required:"true"`
	ChatBaseURL string `envconfig:"CHAT_BASE_URL"`
	ChatPath    string `envconfig:"CHAT_PATH" default:"/ai/user/chat"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("ML", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get ML client config: %w", err)
		panic(err)
	}

	return config
}

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

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(config Config) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(strings.TrimSpace(config.BaseURL), "/"),
		httpClient: http.DefaultClient,
	}
}

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
	defer resp.Body.Close()

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