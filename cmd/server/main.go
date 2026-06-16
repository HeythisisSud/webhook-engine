package main

import (
	"context"
	"fmt"

	"github.com/heythisissud/webhook-engine/internal/config"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/heythisissud/webhook-engine/internal/worker"
	"github.com/heythisissud/webhook-engine/internal/api"

	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/heythisissud/webhook-engine/internal/db/generated"
	"github.com/heythisissud/webhook-engine/internal/poller"

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
	
	queries:=db.New(pool)
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
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"})

	p := poller.NewPoller(queries, asynqClient )
	go p.Start(context.Background())

	r:=chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request){

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
	http.ListenAndServe(":8080", r)



	


	_ = pool // suppress unused warning for now
}
