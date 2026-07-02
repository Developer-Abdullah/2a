package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

type installRepo struct {
	db *postgres.DB
}

func NewInstallRepository(db *postgres.DB) *installRepo {
	return &installRepo{db: db}
}

// GetSignedVersion returns the OTA install data for a specific, successfully signed version.
func (r *installRepo) GetSignedVersion(ctx context.Context, tenantID uuid.UUID, versionID string) (*domain.InstallTarget, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	query := fmt.Sprintf(`
		SELECT
			a.name AS app_name,
			a.bundle_identifier AS bundle_id,
			v.version AS version,
			COALESCE(v.signed_ipa_s3_key, '') AS signed_key,
			COALESCE(a.icon_s3_key, '') AS icon_key
		FROM %s.application_versions v
		JOIN %s.applications a ON a.id = v.application_id
		WHERE v.id = $1 AND v.signing_status = 'signed' AND v.signed_ipa_s3_key IS NOT NULL
	`, schema, schema)

	var target domain.InstallTarget
	if err := r.db.GetContext(ctx, &target, query, versionID); err != nil {
		return nil, err
	}
	return &target, nil
}

// GetLatestSignedVersionID returns the newest signed version for an app (used for install info).
func (r *installRepo) GetLatestSignedVersionID(ctx context.Context, tenantID uuid.UUID, appID string) (string, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf(`
		SELECT v.id::text
		FROM %s.application_versions v
		WHERE v.application_id = $1 AND v.signing_status = 'signed'
		ORDER BY v.created_at DESC
		LIMIT 1
	`, postgres.QuoteIdentifier(schemaName))

	var versionID string
	if err := r.db.GetContext(ctx, &versionID, query, appID); err != nil {
		return "", err
	}
	return versionID, nil
}

func (r *installRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}
