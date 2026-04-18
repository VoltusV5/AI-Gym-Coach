package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_models_simpleconnection "sport_app/internal/core/models/simple_connection"
	simplesql "sport_app/internal/core/models/simple_sql"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_server "sport_app/internal/core/transport/http/server"
	"sport_app/internal/features/handlers"
	"sport_app/internal/features/nutrition"
	"syscall"

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
		fmt.Println("Failed to init application logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("starting sport_app backend")

	dbConfig, err := core_models_simpleconnection.NewConfig()
	if err != nil {
		logger.Fatal("failed to get postgres config", zap.Error(err))
	}

	pool, err := core_models_simpleconnection.NewConnectionPool(ctx, dbConfig)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	auth.InitJWTFromEnv()

	handlers.InitDBPool(pool)

	if err := simplesql.EnsureExercisesSeeded(ctx, pool); err != nil {
		panic(err)
	}

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestID(),
		core_http_middleware.CORS(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Panic(),
		core_http_middleware.Trace(),
	)

	httpServer.RegisterLegacyRoute("POST", "/auth/guest", http.HandlerFunc(handlers.GuestHandler))

	protectedProfile := core_http_middleware.Protect()(http.HandlerFunc(handlers.ProfileHandler))
	httpServer.RegisterLegacyRoute("POST", "/profile", protectedProfile)

	v1Router := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)

	protectedGenerate := core_http_middleware.Protect()(http.HandlerFunc(handlers.ResponceGenerateHandler))
	protectedComplete := core_http_middleware.Protect()(http.HandlerFunc(handlers.WorkoutCompleteHandler))

	v1Router.RegisterRoutes(
		core_http_server.NewRoute("POST", "/plans/generate", protectedGenerate),
		core_http_server.NewRoute("POST", "/workouts/complete", protectedComplete),
	)

	nutritionService := nutrition.NewService(pool.Pool)
	nutritionService.RegisterRoutes(func(method, path string, handler http.Handler) {
		v1Router.RegisterRoutes(core_http_server.NewRoute(method, path, handler))
	})

	httpServer.RegisterAPIRouters(v1Router)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
