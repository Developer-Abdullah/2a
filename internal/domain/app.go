package domain

import (
	"github.com/google/uuid"
	"time"
)

type AppCategory string
type Application struct {
	ID               uuid.UUID   `db:"id" json:"id"`
	BundleIdentifier string      `db:"bundle_identifier" json:"bundle_identifier"`
	Name             string      `db:"name" json:"name"`
	Category         AppCategory `db:"category" json:"category"`
	Description      string      `db:"description" json:"description"`
	Features         []string    `db:"features" json:"features"`
	IconS3Key        *string     `db:"icon_s3_key" json:"icon_s3_key,omitempty"`
	IsPublished      bool        `db:"is_published" json:"is_published"`
	CreatedAt        time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time   `db:"updated_at" json:"updated_at"`
}
type ApplicationVersion struct {
	ID             uuid.UUID `db:"id" json:"id"`
	ApplicationID  uuid.UUID `db:"application_id" json:"application_id"`
	Version        string    `db:"version" json:"version"`
	BuildNumber    string    `db:"build_number" json:"build_number"`
	SizeBytes      int64     `db:"size_bytes" json:"size_bytes"`
	RawIPAS3Key    *string   `db:"raw_ipa_s3_key" json:"raw_ipa_s3_key,omitempty"`
	SignedIPAS3Key *string   `db:"signed_ipa_s3_key" json:"signed_ipa_s3_key,omitempty"`
	ManifestS3Key  *string   `db:"manifest_s3_key" json:"manifest_s3_key,omitempty"`
	SigningStatus  string    `db:"signing_status" json:"signing_status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
type AppItemDTO struct {
	ID               string `db:"id" json:"id"`
	Name             string `db:"name" json:"name"`
	Category         string `db:"category" json:"category"`
	IconURL          string `db:"icon_url" json:"icon_url"`
	BundleIdentifier string `db:"bundle_identifier" json:"bundle_identifier"`
	IsPublished      bool   `db:"is_published" json:"is_published"`
	CreatedAt        string `db:"created_at" json:"created_at"`
}
