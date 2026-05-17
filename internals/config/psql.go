package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPsql() (*pgxpool.Pool, error) {
	values := make([]any, 0, 5)
	values = append(values, os.Getenv("DB_USER"))
	values = append(values, os.Getenv("DB_PASS"))
	values = append(values, os.Getenv("DB_HOST"))
	values = append(values, os.Getenv("DB_PORT"))
	values = append(values, os.Getenv("DB_NAME"))

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", values...)
	return pgxpool.New(context.Background(), connStr)
	// postgresql://USER:PASS@HOST:PORT/DBNAME?sslmode=disable

}
