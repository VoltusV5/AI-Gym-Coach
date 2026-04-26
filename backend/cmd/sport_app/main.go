package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_postgres_pool "sport_app/internal/core/repository/postgres/pool"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/features/mlclient"
	users_postgres_repository "sport_app/internal/features/users/repository/postgres"
	users_service "sport_app/internal/features/users/service"
	users_transport_http "sport_app/internal/features/users/transport/http"

	"go.uber.org/zap"
)

func main() {
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

	logger.Debug("initializing postgres connection pool")
	pool, err := core_postgres_pool.NewConnectionPool(
		ctx,
		core_postgres_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing JWT")
	jwt := core_auth.NewJWT(core_auth.NewConfigMust())

	logger.Debug("initializing ML client")
	mlClient := mlclient.NewClient(mlclient.NewConfigMust())

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository, mlClient, jwt)

	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService, jwt)

	if err := usersService.EnsureExercisesSeeded(ctx); err != nil {
		logger.Fatal("failed to ensure exercises seeded", zap.Error(err))
	}

	logger.Debug("initializing HTTP Server")
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestID(),
		core_http_middleware.CORS(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Panic(),
		core_http_middleware.Trace(),
	)

	apiVersionRouter := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(usersTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
