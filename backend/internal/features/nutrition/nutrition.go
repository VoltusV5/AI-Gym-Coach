package nutrition

import (
	"net/http"

	core_auth "sport_app/internal/core/auth"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	nutrition_repo "sport_app/internal/features/nutrition/repository/postgres"
	nutrition_service "sport_app/internal/features/nutrition/service"
	nutrition_http "sport_app/internal/features/nutrition/transport/http"
)

// Service объединяет зависимости компонента и предоставляет его операции.
type Service struct {
	Repo    *nutrition_repo.Repository
	Service *nutrition_service.NutritionService
	Handler *nutrition_http.NutritionHTTPHandler
}

// NewService создаёт компонент с необходимыми зависимостями.
func NewService(pool core_postgres_pool.Pool) *Service {
	repo := nutrition_repo.NewRepository(pool)
	svc := nutrition_service.NewNutritionService(repo)
	handler := nutrition_http.NewNutritionHTTPHandler(svc)

	return &Service{
		Repo:    repo,
		Service: svc,
		Handler: handler,
	}
}

// RegisterRoutes регистрирует HTTP-маршруты компонента с проверкой JWT на защищённых операциях.
func (s *Service) RegisterRoutes(
	jwt *core_auth.JWT,
	register func(method, path string, handler http.Handler),
) {
	s.Handler.RegisterRoutes(jwt, register)
}
