package api

import (
	"encoding/json"

	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	db "github.com/heythisissud/webhook-engine/internal/db/generated"
	"github.com/heythisissud/webhook-engine/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

type WebhookHandler struct {
	queries *db.Queries
}

func NewWebhookHandler(queries *db.Queries) *WebhookHandler {
	return &WebhookHandler{queries: queries}
}

func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWebhookRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)

	}

	res, err := h.queries.CreateWebhook(r.Context(), db.CreateWebhookParams{
		ClientID:   req.ClientID,
		TargetUrl:  req.TargetURL,
		Secret:     req.Secret,
		EventTypes: req.EventTypes,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)

}

func (h *WebhookHandler) GetWebhookByClientId(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")

	res, err := h.queries.GetWebhooksByClientID(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(res)
}

func (h *WebhookHandler) GetWebhookById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedID, er := uuid.Parse(id)
	if er != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	res, err := h.queries.GetWebhook(r.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})

	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)

	}

	json.NewEncoder(w).Encode(res)

}

func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsedID, er := uuid.Parse(id)
	if er != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err := h.queries.DeleteWebhook(r.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)

	}

	w.WriteHeader(202)

}
