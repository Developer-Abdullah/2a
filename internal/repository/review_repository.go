package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

// ErrInvalidReviewTransition is returned when an approve/reject is attempted on a version that is
// not currently pending_review (already reviewed, or missing).
var ErrInvalidReviewTransition = errors.New("version is not pending review")

type reviewRepo struct {
	db *postgres.DB
}

func NewReviewRepository(db *postgres.DB) *reviewRepo {
	return &reviewRepo{db: db}
}

// ListPendingReviews aggregates the pending_review versions across every active tenant. At the
// roadmap's target scale (< 100 tenants) a per-schema loop is simpler and safer than a generated
// cross-schema UNION; each tenant's rows are tagged with its slug.
func (r *reviewRepo) ListPendingReviews(ctx context.Context) ([]domain.PendingReviewDTO, error) {
	type tenantRow struct {
		Slug   string `db:"slug"`
		Schema string `db:"schema_name"`
	}
	var tenants []tenantRow
	if err := r.db.SelectContext(ctx, &tenants, `
		SELECT slug, schema_name FROM public.tenants WHERE status = 'active' ORDER BY slug
	`); err != nil {
		return nil, err
	}

	out := []domain.PendingReviewDTO{}
	for _, tn := range tenants {
		schema := postgres.QuoteIdentifier(tn.Schema)
		query := fmt.Sprintf(`
			SELECT
				v.id::text AS version_id,
				a.id::text AS app_id,
				a.name AS app_name,
				a.bundle_identifier AS bundle_id,
				v.version AS version,
				v.build_number AS build_number,
				v.size_bytes AS size_bytes,
				COALESCE(v.raw_ipa_s3_key, '') AS raw_ipa_key,
				v.created_at::text AS created_at
			FROM %s.application_versions v
			JOIN %s.applications a ON a.id = v.application_id
			WHERE v.review_status = 'pending_review'
			ORDER BY v.created_at ASC
		`, schema, schema)

		var rows []domain.PendingReviewDTO
		if err := r.db.SelectContext(ctx, &rows, query); err != nil {
			return nil, fmt.Errorf("list pending for %s: %w", tn.Slug, err)
		}
		for i := range rows {
			rows[i].TenantSlug = tn.Slug
		}
		out = append(out, rows...)
	}
	return out, nil
}

// ApproveVersion transitions a version pending_review -> approved and returns the linked signing
// job so the caller can enqueue it. The UPDATE is conditional on the current state, so a version
// that is already reviewed yields ErrInvalidReviewTransition rather than a silent no-op.
func (r *reviewRepo) ApproveVersion(ctx context.Context, tenantSlug, versionID, reviewerID string) (*domain.ApprovalResult, error) {
	tenantID, schemaName, err := r.tenantForSlug(ctx, tenantSlug)
	if err != nil {
		return nil, err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.application_versions
		SET review_status = 'approved', reviewed_by = NULLIF($1, '')::uuid, reviewed_at = NOW()
		WHERE id = $2::uuid AND review_status = 'pending_review'
	`, schema), reviewerID, versionID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrInvalidReviewTransition
	}

	var out domain.ApprovalResult
	out.VersionID = versionID
	out.TenantID = tenantID
	if err := tx.QueryRowxContext(ctx, fmt.Sprintf(`
		SELECT COALESCE(signing_job_id::text, ''), COALESCE(certificate_id::text, '')
		FROM %s.application_versions WHERE id = $1::uuid
	`, schema), versionID).Scan(&out.JobID, &out.CertID); err != nil {
		return nil, err
	}

	if err := r.writeAudit(ctx, tx, schema, "version.approved", versionID, reviewerID, ""); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &out, nil
}

// RejectVersion transitions a version pending_review -> rejected with a required reason.
func (r *reviewRepo) RejectVersion(ctx context.Context, tenantSlug, versionID, reviewerID, reason string) error {
	_, schemaName, err := r.tenantForSlug(ctx, tenantSlug)
	if err != nil {
		return err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.application_versions
		SET review_status = 'rejected', rejection_reason = $1,
			reviewed_by = NULLIF($2, '')::uuid, reviewed_at = NOW()
		WHERE id = $3::uuid AND review_status = 'pending_review'
	`, schema), reason, reviewerID, versionID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrInvalidReviewTransition
	}

	if err := r.writeAudit(ctx, tx, schema, "version.rejected", versionID, reviewerID, reason); err != nil {
		return err
	}
	return tx.Commit()
}

// writeAudit records a review action in the tenant's tamper-proof audit_logs. The reviewer is a
// platform owner (public.platform_admins), not a tenant admin_users row, so admin_id is left NULL
// and the reviewer id travels in new_value.
func (r *reviewRepo) writeAudit(ctx context.Context, tx sqlxExecer, schema, action, versionID, reviewerID, reason string) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s.audit_logs (admin_id, action, resource_type, resource_id, new_value)
		VALUES (NULL, $1, 'application_version', $2, jsonb_build_object('reviewer_id', $3, 'reason', $4))
	`, schema), action, versionID, reviewerID, reason)
	return err
}

func (r *reviewRepo) tenantForSlug(ctx context.Context, slug string) (id, schemaName string, err error) {
	var row struct {
		ID     string `db:"id"`
		Schema string `db:"schema_name"`
	}
	err = r.db.GetContext(ctx, &row, `SELECT id::text, schema_name FROM public.tenants WHERE slug = $1`, slug)
	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("tenant %q not found", slug)
	}
	return row.ID, row.Schema, err
}

// sqlxExecer is the minimal surface writeAudit needs from a transaction.
type sqlxExecer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
