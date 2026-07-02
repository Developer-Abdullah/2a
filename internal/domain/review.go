package domain

// ReviewStatus is the manual-review state of an application version. Every version starts life as
// ReviewPending and must be approved by the platform owner before its signing job may run.
type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending_review"
	ReviewApproved ReviewStatus = "approved"
	ReviewRejected ReviewStatus = "rejected"
)

// CanTransitionTo reports whether moving from the receiver state to `to` is a legal review
// transition. Only pending_review -> approved and pending_review -> rejected are allowed; approved
// and rejected are terminal. This is the single source of truth mirrored by the DB trigger in
// migration 017 and used as a pre-check by the service layer.
func (s ReviewStatus) CanTransitionTo(to ReviewStatus) bool {
	if s != ReviewPending {
		return false
	}
	return to == ReviewApproved || to == ReviewRejected
}

// PendingReviewDTO is one entry in the platform owner's cross-tenant review queue.
type PendingReviewDTO struct {
	TenantSlug  string `db:"tenant_slug" json:"tenant_slug"`
	VersionID   string `db:"version_id" json:"version_id"`
	AppID       string `db:"app_id" json:"app_id"`
	AppName     string `db:"app_name" json:"app_name"`
	BundleID    string `db:"bundle_id" json:"bundle_id"`
	Version     string `db:"version" json:"version"`
	BuildNumber string `db:"build_number" json:"build_number"`
	SizeBytes   int64  `db:"size_bytes" json:"size_bytes"`
	RawIPAKey   string `db:"raw_ipa_key" json:"raw_ipa_key"`
	CreatedAt   string `db:"created_at" json:"created_at"`
}

// ApprovalResult carries the data the controller needs to enqueue a signing job after approval.
type ApprovalResult struct {
	JobID     string
	VersionID string
	CertID    string
	TenantID  string // tenant UUID — the signing worker resolves the schema by id
}
