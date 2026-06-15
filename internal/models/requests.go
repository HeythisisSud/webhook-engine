package models



type CreateWebhookRequest struct {
    ClientID   string   `json:"client_id"`
    TargetURL  string   `json:"target_url"`
    Secret     string   `json:"secret"`
    EventTypes []string `json:"event_types"`
}