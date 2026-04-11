package main

import (
	"context"
	"fmt"
	auth "sport_app/internal/core/auth"
	simpleconnection "sport_app/internal/core/models/simple_connection"
	simplesql "sport_app/internal/core/models/simple_sql"
	"sport_app/internal/features/handlers"
)

func main() {
	auth.InitJWTFromEnv()

	config, err := simpleconnection.NewConfig()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	pool, err := simpleconnection.NewConnectionPool(ctx, config)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := simplesql.EnsureExercisesSeeded(ctx, pool); err != nil {
		panic(err)
	}

	fmt.Println("Успешно!")
	if err := handlers.StartHTTPServer(pool); err != nil {
		fmt.Println("Failed to start HTTP server")
		return
	}
}
