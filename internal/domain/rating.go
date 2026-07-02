package domain

import (
	"time"

	"github.com/google/uuid"
)

type PlatformRating struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    *uuid.UUID `db:"user_id" json:"user_id,omitempty"`
	Rating    int16      `db:"rating" json:"rating"`
	Comment   *string    `db:"comment" json:"comment,omitempty"`
	IPHash    string     `db:"ip_hash" json:"-"`
	DeviceID  *uuid.UUID `db:"device_id" json:"device_id,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// RatingSummary holds the aggregate metrics for GET /v1/ratings/summary.
// R1..R5 are the per-star counts used to build the breakdown map.
type RatingSummary struct {
	Average float64 `db:"average"`
	Count   int     `db:"count"`
	R1      int     `db:"r1"`
	R2      int     `db:"r2"`
	R3      int     `db:"r3"`
	R4      int     `db:"r4"`
	R5      int     `db:"r5"`
}