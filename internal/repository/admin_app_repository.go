package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

type adminAppRepo struct {
	db *postgres.DB
}

func NewAdminAppRepository(db *postgres.DB) *adminAppRepo {
	return &adminAppRepo{db: db}
}

// GetAppByID returns an app regardless of publish state (admin view).
func (r *adminAppRepo) GetAppByID(ctx context.Context, tenantID uuid.UUID, appID string) (*domain.AppItemDTO, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT id::text AS id, name, category::text AS category,
			COALESCE(icon_s3_key, '') AS icon_url, bundle_identifier, is_published, created_at::text AS created_at
		FROM %s.applications WHERE id = $1
	`, postgres.QuoteIdentifier(schemaName))

	var app domain.AppItemDTO
	if err := r.db.GetContext(ctx, &app, query, appID); err != nil {
		return nil, err
	}
	return &app, nil
}

// CreateApp inserts a new application and returns its id.
func (r *adminAppRepo) CreateApp(ctx context.Context, tenantID uuid.UUID, bundleID, name, category, description string, features []string, publish bool) (string, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if category == "" {
		category = "other"
	}
	// The app_category enum lives in the tenant schema, so the cast must be schema-qualified —
	// search_path does not include the tenant schema on this pooled connection.
	query := fmt.Sprintf(`
		INSERT INTO %[1]s.applications (bundle_identifier, name, category, description, features, is_published)
		VALUES ($1, $2, $3::%[1]s.app_category, NULLIF($4, ''), $5::text[], $6)
		RETURNING id::text
	`, postgres.QuoteIdentifier(schemaName))

	var id string
	if err := r.db.QueryRowxContext(ctx, query, bundleID, name, category, description, encodePGTextArray(features), publish).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

// ListApps returns every app for the tenant (published or not) for the admin dashboard.
func (r *adminAppRepo) ListApps(ctx context.Context, tenantID uuid.UUID) ([]domain.AppItemDTO, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT id::text AS id, name, category::text AS category,
			COALESCE(icon_s3_key, '') AS icon_url, bundle_identifier, is_published, created_at::text AS created_at
		FROM %s.applications ORDER BY created_at DESC
	`, postgres.QuoteIdentifier(schemaName))

	var items []domain.AppItemDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AppItemDTO{}
	}
	return items, nil
}

// ListVersions returns the version history (with signing status) for an app.
func (r *adminAppRepo) ListVersions(ctx context.Context, tenantID uuid.UUID, appID string) ([]domain.VersionDTO, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT id::text AS id, version, build_number, signing_status::text AS signing_status,
			review_status::text AS review_status, COALESCE(rejection_reason, '') AS rejection_reason,
			COALESCE(signed_ipa_s3_key, '') AS signed_key, size_bytes, created_at::text AS created_at
		FROM %s.application_versions WHERE application_id = $1 ORDER BY created_at DESC
	`, postgres.QuoteIdentifier(schemaName))

	var items []domain.VersionDTO
	if err := r.db.SelectContext(ctx, &items, query, appID); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.VersionDTO{}
	}
	return items, nil
}

// ActiveCertID returns the tenant's active signing certificate id (required to enqueue a job).
func (r *adminAppRepo) ActiveCertID(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var certID string
	err := r.db.GetContext(ctx, &certID, `
		SELECT id::text FROM public.tenant_certificates
		WHERE tenant_id = $1 AND is_active = true AND revoked_at IS NULL
		ORDER BY added_at DESC LIMIT 1
	`, tenantID)
	if err != nil {
		return "", err
	}
	return certID, nil
}

// CreateVersionWithJob atomically inserts a pending version + queued signing job and links them.
func (r *adminAppRepo) CreateVersionWithJob(ctx context.Context, tenantID uuid.UUID, appID, version, build, rawKey string, sizeBytes int64, certID string) (string, string, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return "", "", err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback()

	var versionID string
	err = tx.QueryRowxContext(ctx, fmt.Sprintf(`
		INSERT INTO %s.application_versions (application_id, version, build_number, size_bytes, raw_ipa_s3_key, signing_status, certificate_id)
		VALUES ($1::uuid, $2, $3, $4, NULLIF($5, ''), 'pending', NULLIF($6, '')::uuid)
		RETURNING id::text
	`, schema), appID, version, build, sizeBytes, rawKey, certID).Scan(&versionID)
	if err != nil {
		return "", "", err
	}

	var jobID string
	err = tx.QueryRowxContext(ctx, fmt.Sprintf(`
		INSERT INTO %s.signing_jobs (version_id, certificate_id, status)
		VALUES ($1::uuid, $2::uuid, 'queued')
		RETURNING id::text
	`, schema), versionID, certID).Scan(&jobID)
	if err != nil {
		return "", "", err
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.application_versions SET signing_job_id = $1::uuid WHERE id = $2::uuid
	`, schema), jobID, versionID); err != nil {
		return "", "", err
	}

	if err := tx.Commit(); err != nil {
		return "", "", err
	}
	return versionID, jobID, nil
}

func (r *adminAppRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}

// encodePGTextArray renders a Go slice into a PostgreSQL array literal (e.g. {"a","b"}).
func encodePGTextArray(items []string) string {
	if len(items) == 0 {
		return "{}"
	}
	escaped := make([]string, len(items))
	for i, item := range items {
		s := strings.ReplaceAll(item, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		escaped[i] = `"` + s + `"`
	}
	return "{" + strings.Join(escaped, ",") + "}"
}
