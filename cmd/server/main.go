package main

import (
	"context"
	"fmt"

	"github.com/heythisissud/webhook-engine/internal/api"
	"github.com/heythisissud/webhook-engine/internal/config"
	"github.com/heythisissud/webhook-engine/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/heythisissud/webhook-engine/internal/db/generated"
	"github.com/heythisissud/webhook-engine/internal/poller"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"log"
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

	queries := db.New(pool)
	// create asynq server (this is the worker, reads from redis)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{Concurrency: 10},
	)

	// register which function handles which task type
	mux := asynq.NewServeMux()
	w := worker.NewWorkerQuery(queries)
	mux.HandleFunc(worker.TypeWebhookDelivery, w.HandleWebhookDelivery)

	// start worker in a goroutine so it doesn't block
	go srv.Run(mux)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	p := poller.NewPoller(queries, asynqClient)
	go p.Start(context.Background())

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("ok"))
	})
	webhookHandler := api.NewWebhookHandler(queries)

	r.Post("/webhooks", webhookHandler.CreateWebhook)
	r.Get("/webhooks", webhookHandler.GetWebhookByClientId)
	r.Get("/webhooks/{id}", webhookHandler.GetWebhookById)
	r.Delete("/webhooks/{id}", webhookHandler.DeleteWebhook)
	eventHandler := api.NewEventHandler(queries, pool)
	r.Post("/events", eventHandler.IngestEvent)
	// // event routes
	// r.Post("/events", nil) // ingest an event

	// start server
	// create http server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// listen for CTRL+C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// start server in goroutine
	go func() {
		log.Println("server starting on port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error:", err)
		}
	}()

	// block until signal received
	<-quit
	log.Println("shutting down...")

	// give in-flight requests 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	pool.Close()
	asynqClient.Close()

	log.Println("shutdown complete")

	_ = pool // suppress unused warning for now
}
