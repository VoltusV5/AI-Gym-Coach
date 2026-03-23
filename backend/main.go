package main

import (
	"fmt"
	"sport_app/auth"
	"sport_app/handlers"
	simpleconnection "sport_app/models/simple_connection"
	simplesql "sport_app/models/simple_sql"
)

func main() {
	auth.InitJWTFromEnv()

	simpleconnection.CheckConnection()
	defer simpleconnection.Close()

	if err := simplesql.CreateTableUsers(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}

	if err := simplesql.CreateTableProfile(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}

	if err := simplesql.CreateTableData(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}

	if err := simplesql.CreateTableExercises(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}

	if err := simplesql.CreateTableProgram(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}

	if err := simplesql.EnsureExercisesSeeded(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}

	fmt.Println("Успешно!")
	if err := handlers.StartHTTPServer(); err != nil {
		fmt.Println("Failed to start HTTP server")
		return
	}
}
