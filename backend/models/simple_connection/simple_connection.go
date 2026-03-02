package simpleconnection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

var Ctx = context.Background()
var Conn *pgx.Conn

func CheckConnection() {
	connString := os.Getenv("MY_KEY")
	var err error
	Conn, err = pgx.Connect(Ctx, connString)
	if err != nil {
		panic(err)
	}
	if err := Conn.Ping(Ctx); err != nil {
		fmt.Println("Пинг базы данных не удался")
	}
}

func Close() {
	if Conn != nil {
		Conn.Close(Ctx)
	}
}
