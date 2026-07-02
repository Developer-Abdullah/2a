package domain

import (
	"github.com/google/uuid"
	"time"
)

type SigningJobStatus string

const (
	SigningStatusQueued     SigningJobStatus = "queued"
	SigningStatusProcessing SigningJobStatus = "processing"
	SigningStatusCompleted  SigningJobStatus = "completed"
	SigningStatusFailed     SigningJobStatus = "failed"
)

type SigningJob struct {
	ID            uuid.UUID        `db:"id" json:"id"`
	VersionID     uuid.UUID        `db:"version_id" json:"version_id"`
	CertificateID uuid.UUID        `db:"certificate_id" json:"certificate_id"`
	Status        SigningJobStatus `db:"status" json:"status"`
	Priority      int              `db:"priority" json:"priority"`
	QueuedAt      time.Time        `db:"queued_at" json:"queued_at"`
	RetryCount    int              `db:"retry_count" json:"retry_count"`
}
type SigningJobStatusDTO struct {
	Status      string `db:"status" json:"status"`
	ProgressPct int    `db:"progress_pct" json:"progress_pct"`
	ErrorMsg    string `db:"error_message" json:"error_message,omitempty"`
}
