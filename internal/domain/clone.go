package domain

import (
	"time"

	"github.com/google/uuid"
)

type CloneStatus string

const (
	CloneStatusPending CloneStatus = "pending"
	CloneStatusSigning CloneStatus = "signing"
	CloneStatusActive  CloneStatus = "active"
	CloneStatusRevoked CloneStatus = "revoked"
)

type CloneRecord struct {
	ID                uuid.UUID   `db:"id" json:"id"`
	OriginalAppID     uuid.UUID   `db:"original_app_id" json:"original_app_id"`
	OriginalVersionID uuid.UUID   `db:"original_version_id" json:"original_version_id"`
	ClonedBundleID    string      `db:"cloned_bundle_id" json:"cloned_bundle_id"`
	UserID            uuid.UUID   `db:"user_id" json:"user_id"`
	DeviceID          uuid.UUID   `db:"device_id" json:"device_id"`
	SigningJobID      *uuid.UUID  `db:"signing_job_id" json:"signing_job_id,omitempty"`
	PINVerifiedAt     *time.Time  `db:"pin_verified_at" json:"pin_verified_at,omitempty"`
	Status            CloneStatus `db:"status" json:"status"`
	CreatedAt         time.Time   `db:"created_at" json:"created_at"`
}