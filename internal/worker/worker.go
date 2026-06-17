package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	db "github.com/heythisissud/webhook-engine/internal/db/generated"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

const (
	TypeWebhookDelivery = "webhook:delivery"
)

type WebhookDeliveryPayload struct {
	OutboxId   string `json:"outbox_id"`
	WebhookUrl string `json:"webhook_url"`
	Payload    []byte `json:"payload"`
    Secret     string `json:"secret"`

}

type WorkerQuery struct {
	query *db.Queries
}

func NewWorkerQuery(query *db.Queries) *WorkerQuery {
	return &WorkerQuery{
		query: query,
	}
}

func NewWebhookDeliveryTask(OutboxId string, WebhookUrl string, Payload []byte,Secret string) (*asynq.Task, error) {
	payload, err := json.Marshal(WebhookDeliveryPayload{
		OutboxId:   OutboxId,
		WebhookUrl: WebhookUrl,
		Payload:    Payload,
		Secret: Secret,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeWebhookDelivery, payload), nil
}

func signPayload(payload []byte, secret string) string {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (w *WorkerQuery) HandleWebhookDelivery(ctx context.Context, t *asynq.Task) error {
	var p WebhookDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)

	}
	// 2. make the HTTP POST to the target URL
	// sign the payload
	signature := signPayload(p.Payload, p.Secret)
	
	// create request manually
	req, err := http.NewRequest("POST", p.WebhookUrl, bytes.NewBuffer(p.Payload))
	if err != nil {
	    return fmt.Errorf("failed to create request: %v", err)
	}
	
	// set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	
	// send it
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
	    log.Println("delivery error:", err)
	    return fmt.Errorf("http delivery failed: %v", err)
	}
	defer resp.Body.Close()
	log.Println("delivery response status:", resp.StatusCode)
	// prepare values
	statusCode := int32(resp.StatusCode)
	success := resp.StatusCode >= 200 && resp.StatusCode <= 299

	// read response body
	body, _ := io.ReadAll(resp.Body)

	// write to delivery_logs
	parsedID, er := uuid.Parse(p.OutboxId)
	if er != nil {
		log.Printf("Error parsing UUID: %v", er)
		return fmt.Errorf("failed to parse UUID: %w", er)
	}
	retryCount, _ := asynq.GetRetryCount(ctx)

	w.query.CreatedDeliveryLog(ctx, db.CreatedDeliveryLogParams{
		OutboxID:      pgtype.UUID{Bytes: parsedID, Valid: true},
		AttemptNumber: int32(retryCount+1),
		StatusCode:    pgtype.Int4{Int32: statusCode, Valid: true},
		ResponseBody:  pgtype.Text{String: string(body), Valid: true},
		ErrorMessage:  pgtype.Text{},
		Success:       success,
	})

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		w.query.UpdateOutboxStatus(ctx, db.UpdateOutboxStatusParams{
			ID:     pgtype.UUID{Bytes: parsedID, Valid: true},
			Status: "failed",
		})
		return fmt.Errorf("non-2xx status code: %d", resp.StatusCode)
	}

	err = w.query.UpdateOutboxStatus(ctx, db.UpdateOutboxStatusParams{
		ID:     pgtype.UUID{Bytes: [16]byte(parsedID), Valid: true},
		Status: "delivered",
	})
	if err != nil {
		log.Println("error updating outbox status:", err)
	}

	// 3. if non-2xx, return error so asynq retries
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("non-2xx status code: %d", resp.StatusCode)
	}

	log.Printf("delivered to %s — status %d", p.WebhookUrl, resp.StatusCode)

	return nil
}
