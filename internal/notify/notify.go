package notify

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// TypePush is the asynq task type consumed by the notifier worker on the "notifications" queue.
const TypePush = "notification:push"

type PushPayload struct {
	NotificationID string `json:"notification_id"`
	TenantID       string `json:"tenant_id"`
	Title          string `json:"title"`
	Body           string `json:"body"`
}

func NewPushTask(p PushPayload) (*asynq.Task, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypePush, b), nil
}

// TypeEmail is the asynq task type for transactional emails (e.g. delivering activation codes after
// a paid order), consumed by the notifier worker on the "notifications" queue.
const TypeEmail = "email:send"

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewEmailTask(p EmailPayload) (*asynq.Task, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEmail, b), nil
}
