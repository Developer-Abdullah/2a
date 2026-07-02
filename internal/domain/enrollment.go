package domain

import (
	"time"

	"github.com/google/uuid"
)

type UDIDEnrollmentSession struct {
	ID           uuid.UUID   `db:"id" json:"id"`
	TenantID     uuid.UUID   `db:"tenant_id" json:"tenant_id"`
	OneTimeToken string      `db:"one_time_token" json:"one_time_token"`
	UDIDHash     *string     `db:"udid_hash" json:"udid_hash,omitempty"`
	DeviceType   *DeviceType `db:"device_type" json:"device_type,omitempty"`
	IPAddress    string      `db:"ip_address" json:"-"`
	Completed    bool        `db:"completed" json:"completed"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
	ExpiresAt    time.Time   `db:"expires_at" json:"expires_at"`
}