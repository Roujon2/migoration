package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func ConnectDB(connStr string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}