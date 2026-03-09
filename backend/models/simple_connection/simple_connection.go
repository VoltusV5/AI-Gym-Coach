package simpleconnection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Ctx = context.Background()
var Conn *pgxpool.Pool

func CheckConnection() {
	connString := os.Getenv("MY_KEY")
	var err error
	Conn, err = pgxpool.New(Ctx, connString)
	if err != nil {
		panic(err)
	}
	if err := Conn.Ping(Ctx); err != nil {
		fmt.Println("Пинг базы данных не удался")
	}
}

func Close() {
	if Conn != nil {
		Conn.Close()
	}
}
