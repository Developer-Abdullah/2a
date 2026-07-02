package domain

import "github.com/google/uuid"

type DeviceType string
type User struct {
	ID          uuid.UUID `db:"id" json:"id"`
	DisplayName string    `db:"display_name" json:"display_name"`
}
type Device struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	DeviceType DeviceType `db:"device_type" json:"device_type"`
	IsRevoked  bool       `db:"is_revoked" json:"is_revoked"`
}
