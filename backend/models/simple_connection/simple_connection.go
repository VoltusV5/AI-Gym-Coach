package simpleconnection

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func CheckConnection(ctx context.Context) (*pgx.Conn, error) {
	connString := os.Getenv("MY_KEY")
	return pgx.Connect(ctx, connString)
}
