package users_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"sport_app/internal/features/mlclient"
	users_service "sport_app/internal/features/users/service"

	"github.com/stretchr/testify/require"
)

func TestGeneratePlanUsesConfiguredDeadline(t *testing.T) {
	observed := make(chan time.Time, 1)
	stop := errors.New("тест завершён после проверки срока")
	ml := &mockMLClient{GeneratePlanFunc: func(ctx context.Context, _ any) (*mlclient.Plan, error) {
		deadline, _ := ctx.Deadline()
		observed <- deadline
		return nil, stop
	}}
	svc := users_service.NewUsersService(&mockUsersRepository{}, ml, nil,
		users_service.WithPlanGenerationTimeout(time.Minute))
	started := time.Now()
	_, err := svc.GeneratePlan(context.Background(), "user-1")
	finished := time.Now()
	require.ErrorIs(t, err, stop)
	deadline := <-observed
	require.False(t, deadline.Before(started.Add(time.Minute)))
	require.False(t, deadline.After(finished.Add(time.Minute)))
}
