package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_pgx_pool "sport_app/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/features/mlclient"
	"sport_app/internal/features/mlclient/aichat"
	"sport_app/internal/features/nutrition"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
	users_service "sport_app/internal/features/users/service"
	users_transport_http "sport_app/internal/features/users/transport/http"
	"syscall"

	"go.uber.org/zap"
)

var (
	timeZone = time.UTC
)

func main() {
	time.Local = timeZone
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("Failed to init application logger: ", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", timeZone))

	logger.Debug("initializing postgres connection pool")
	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing JWT")
	jwt := core_auth.NewJWT(core_auth.NewConfigMust())

	logger.Debug("initializing ML client")
	mlCfg := mlclient.NewConfigMust()
	mlClient := mlclient.NewClient(mlCfg)

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository, mlClient, jwt)

	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService, jwt)

	if err := usersService.EnsureExercisesSeeded(ctx); err != nil {
		logger.Fatal("failed to ensure exercises seeded", zap.Error(err))
	}

	if err := usersService.EnsureAchievements(ctx); err != nil {
		logger.Fatal("failed to ensure achievements", zap.Error(err))
	}

	logger.Debug("initializing HTTP Server")
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestID(),
		core_http_middleware.CORS(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	apiVersionRouter := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(usersTransportHTTP.Routes()...)

	registerV1 := func(method, path string, handler http.Handler) {
		var hf http.HandlerFunc
		if h, ok := handler.(http.HandlerFunc); ok {
			hf = h
		} else {
			hf = func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
			}
		}
		apiVersionRouter.RegisterRoutes(core_http_server.NewRoute(method, path, hf))
	}

	nutrition.NewService(pool).RegisterRoutes(jwt, registerV1)
	aichat.NewService(mlCfg, aichat.NewUsersReader(usersRepository)).RegisterRoutes(jwt, registerV1)

	httpServer.RegisterAPIRouters(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
