package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeAlert  NotificationType = "alert"
	NotificationTypeSilent NotificationType = "silent"
	NotificationTypeUpdate NotificationType = "update"
)

type Notification struct {
	ID             uuid.UUID        `db:"id" json:"id"`
	Title          string           `db:"title" json:"title"`
	Body           string           `db:"body" json:"body"`
	TargetAppID    *uuid.UUID       `db:"target_app_id" json:"target_app_id,omitempty"`
	SentAt         time.Time        `db:"sent_at" json:"sent_at"`
	SentByAdminID  *uuid.UUID       `db:"sent_by_admin_id" json:"sent_by_admin_id,omitempty"`
	Type           NotificationType `db:"type" json:"type"`
	ReadAt         *time.Time       `db:"read_at" json:"read_at,omitempty"` // populated via JOIN
}