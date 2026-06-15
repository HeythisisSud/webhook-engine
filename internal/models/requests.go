package models

import "encoding/json"

type CreateWebhookRequest struct {
	ClientID   string   `json:"client_id"`
	TargetURL  string   `json:"target_url"`
	Secret     string   `json:"secret"`
	EventTypes []string `json:"event_types"`
}

type IngestEventRequest struct {
	ClientID  string          `json:"client_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}
