package main

import (
	"context"
	"fmt"

	"github.com/heythisissud/webhook-engine/internal/config"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/heythisissud/webhook-engine/internal/worker"
)

// --- Package-level declarations (these CANNOT go inside a function) ---





// --- Entry point ---

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

	// create asynq server (this is the worker, reads from redis)
	srv := asynq.NewServer(
	    asynq.RedisClientOpt{Addr: "localhost:6379"},
	    asynq.Config{Concurrency: 10},
	)
	
	// register which function handles which task type
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TypeWebhookDelivery, worker.HandleWebhookDelivery)
	
	// start worker in a goroutine so it doesn't block
	go srv.Run(mux)
	_ = pool // suppress unused warning for now
}
