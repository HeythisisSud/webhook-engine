package api

import (
	"encoding/json"
	"net/http"

	db "github.com/heythisissud/webhook-engine/internal/db/generated"
	"github.com/heythisissud/webhook-engine/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventHandler struct {
	queries *db.Queries
	pool    *pgxpool.Pool // need pool directly for transactions
}

func NewEventHandler(queries *db.Queries, pool *pgxpool.Pool) *EventHandler {
	return &EventHandler{queries: queries, pool: pool}
}

func (h *EventHandler) IngestEvent(w http.ResponseWriter, r *http.Request) {
	var req models.IngestEventRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// start transaction
	tx, err := h.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	// create queries object that uses the transaction
	qtx := db.New(tx)

	// 1. insert event
	event, err := qtx.CreateEvents(ctx, db.CreateEventsParams{
		ClientID:  req.ClientID,
		EventType: req.EventType,
		Payload:   []byte(req.Payload),
	})
	if err != nil {
		http.Error(w, "Error creating event", http.StatusInternalServerError)
		return
	}

	webhooks, err := qtx.GetWebhooksByEventType(ctx, []string{req.EventType})
	if err != nil {
		http.Error(w, "Error fetching webhooks", http.StatusInternalServerError)
		return
	}

	for _, webhook := range webhooks {
		_, err := qtx.CreateOutboxEntry(ctx, db.CreateOutboxEntryParams{
			EventID:   event.ID,
			WebhookID: webhook.ID,
		})
		if err != nil {
			http.Error(w, "Error creating outbox entry", http.StatusInternalServerError)
			return
		}
	}

	// all good — commit
	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}
