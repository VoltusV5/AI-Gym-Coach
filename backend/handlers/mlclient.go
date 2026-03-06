package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Plan struct {
	Muscle_group string
	Area         string
	exercises    string
}

func GeneratePlan(ctx context.Context, reqBody interface{}) (*Plan, error) {
	url_string := os.Getenv("ML_BASE_URL")
	url := url_string + "/generate"
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

	return &plan, nil
}
