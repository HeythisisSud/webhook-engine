package main

import (
	"context"
	"fmt"

	"github.com/heythisissud/webhook-engine/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		fmt.Println("error connecting to db", err)
		return
	}

	fmt.Println("db connected")
	_ = pool // suppress unused warning for now
}
