package database

import (
	"context"
	"crudapi/config"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect() *pgxpool.Pool {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DbUrl)
	if err != nil {
		panic(fmt.Sprintf("failed to create connection pool: %v", err))
	}

	if err := pool.Ping(context.Background()); err != nil {
		panic(fmt.Sprintf("database unreachable: %v", err))
	}

	fmt.Println("Connected to database successfully!")

	return pool
}
