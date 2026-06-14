package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	
)







const (
	TypeWebhookDelivery = "webhook:delivery"
)

type WebhookDeliveryPayload struct {
    OutboxId   string `json:"outbox_id"`
    WebhookUrl string `json:"webhook_url"`
    Payload    []byte `json:"payload"`
}

func NewWebhookDeliveryTask(OutboxId string, WebhookUrl string, Payload []byte) (*asynq.Task, error) {
	payload, err := json.Marshal(WebhookDeliveryPayload{
		OutboxId:  OutboxId,
		WebhookUrl: WebhookUrl,
		Payload:    Payload,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeWebhookDelivery, payload), nil
}

func HandleWebhookDelivery(ctx context.Context, t *asynq.Task) error{
	var p WebhookDeliveryPayload
	if err:= json.Unmarshal(t.Payload(),&p); err!=nil{
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)

	}
	log.Printf("Sending delivery to WebhookUrl:  %s", p.WebhookUrl)


	return nil
}