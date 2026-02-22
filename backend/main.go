package main

import (
	"context"
	"fmt"
	simpleconnection "sport_app/models/simple_connection"
	simplesql "sport_app/models/simple_sql"
)

func main() {
	ctx := context.Background()

	conn, err := simpleconnection.CheckConnection(ctx)
	if err != nil {
		panic(err)
	}

	if err := simplesql.CreateTableUsers(ctx, conn); err != nil {
		panic(err)
	}

	if err := simplesql.CreateTableProfile(ctx, conn); err != nil {
		panic(err)
	}
	fmt.Println("Успешно!")
}
