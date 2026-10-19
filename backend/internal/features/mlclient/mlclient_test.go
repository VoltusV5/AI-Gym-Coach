package mlclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sport_app/internal/features/mlclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMLClient_GeneratePlan_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/plan/user", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		responseJSON := `{
			"тип_сплита": "Upper/Lower",
			"еженедельный_план": [
				{
					"день": "Понедельник",
					"тип_дня": "Верх",
					"упражнения": [
						{
							"группа": "Спина",
							"подгруппа": "Верх спины"
						},
						{
							"группа": "Спина",
							"подгруппа": "Широчайшие спины"
						}
					]
				}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer ts.Close()

	client := mlclient.NewClient(mlclient.Config{
		BaseURL: ts.URL,
		Timeout: 5 * time.Second,
	})

	plan, err := client.GeneratePlan(context.Background(), map[string]any{"goal": "hypertrophy"})
	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.Equal(t, "Upper/Lower", plan.Split_type)
	require.Len(t, plan.Plan_week, 1)
	require.Len(t, plan.Plan_week[0].Exercises, 2)

	// Проверяем замену упражнений подгруппы.
	require.NotNil(t, plan.Plan_week[0].Exercises[0].Sub_group)
	assert.Equal(t, "тяга перед собой широким хватом", *plan.Plan_week[0].Exercises[0].Sub_group)

	require.NotNil(t, plan.Plan_week[0].Exercises[1].Sub_group)
	assert.Equal(t, "тяга сверху узким хватом", *plan.Plan_week[0].Exercises[1].Sub_group)
}

func TestMLClient_GeneratePlan_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := mlclient.NewClient(mlclient.Config{
		BaseURL: ts.URL,
		Timeout: 50 * time.Millisecond, // Короткий таймаут для проверки отмены запроса.
	})

	_, err := client.GeneratePlan(context.Background(), map[string]any{"goal": "strength"})
	require.Error(t, err)
	assert.True(t,
		assert.Contains(t, err.Error(), "context deadline exceeded") ||
			assert.Contains(t, err.Error(), "Client.Timeout exceeded") ||
			assert.Contains(t, err.Error(), "timeout"),
	)
}

func TestMLClient_GeneratePlan_Non200Error(t *testing.T) {
	statusCodes := []int{
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	}

	for _, code := range statusCodes {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
			_, _ = w.Write([]byte(`{"error": "ML inference model unavailable"}`))
		}))

		client := mlclient.NewClient(mlclient.Config{
			BaseURL: ts.URL,
			Timeout: 5 * time.Second,
		})

		_, err := client.GeneratePlan(context.Background(), map[string]any{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ML returned status")

		ts.Close()
	}
}

func TestMLClient_GeneratePlan_MalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer ts.Close()

	client := mlclient.NewClient(mlclient.Config{
		BaseURL: ts.URL,
		Timeout: 5 * time.Second,
	})

	_, err := client.GeneratePlan(context.Background(), map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode response")
}

func TestMLClient_GeneratePlan_EmptyBaseURL(t *testing.T) {
	client := mlclient.NewClient(mlclient.Config{
		BaseURL: "",
	})

	_, err := client.GeneratePlan(context.Background(), map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ML base url is not set")
}
