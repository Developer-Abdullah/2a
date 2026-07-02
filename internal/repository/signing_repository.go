package repository

import (
	"context"
	"database/sql"
	"fmt"

	"platform/internal/domain"
	"platform/internal/signing"
	"platform/internal/store/postgres"
)

type signingRepo struct {
	db *postgres.DB
}

func NewSigningRepository(db *postgres.DB) *signingRepo {
	return &signingRepo{db: db}
}

func (r *signingRepo) GetSigningContext(ctx context.Context, jobID string, tenantID string) (*signing.SigningContext, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	// tenant_certificates lives in the public schema; join across schemas to resolve
	// the encrypted cert assets the signing worker needs for this job.
	query := fmt.Sprintf(`
		SELECT
			a.id::text AS app_id,
			a.bundle_identifier AS app_bundle_id,
			a.name AS app_name,
			v.version AS app_version,
			COALESCE(v.raw_ipa_s3_key, '') AS raw_ipa_s3_key,
			j.certificate_id::text AS cert_id,
			COALESCE(c.s3_key_p12_enc, '') AS p12_key,
			COALESCE(c.s3_key_mobileprovision, '') AS provision_key,
			c.encrypted_password AS encrypted_password
		FROM %s.signing_jobs j
		JOIN %s.application_versions v ON v.id = j.version_id
		JOIN %s.applications a ON a.id = v.application_id
		LEFT JOIN public.tenant_certificates c ON c.id = j.certificate_id
		WHERE j.id = $1
	`, schema, schema, schema)

	var dest struct {
		AppID             string `db:"app_id"`
		AppBundleID       string `db:"app_bundle_id"`
		AppName           string `db:"app_name"`
		AppVersion        string `db:"app_version"`
		RawIPAS3Key       string `db:"raw_ipa_s3_key"`
		CertID            string `db:"cert_id"`
		P12Key            string `db:"p12_key"`
		ProvisionKey      string `db:"provision_key"`
		EncryptedPassword []byte `db:"encrypted_password"`
	}
	if err := r.db.GetContext(ctx, &dest, query, jobID); err != nil {
		return nil, err
	}

	return &signing.SigningContext{
		AppID:                dest.AppID,
		AppBundleID:          dest.AppBundleID,
		AppName:              dest.AppName,
		AppVersion:           dest.AppVersion,
		RawIPAS3Key:          dest.RawIPAS3Key,
		CertID:               dest.CertID,
		EncryptedP12S3Key:    dest.P12Key,
		MobileProvisionS3Key: dest.ProvisionKey,
		EncryptedPassword:    dest.EncryptedPassword,
	}, nil
}

func (r *signingRepo) UpdateJobFailed(ctx context.Context, jobID string, tenantID string, errMsg string) error {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		UPDATE %s.signing_jobs
		SET status = 'failed', failed_at = NOW(), error_message = $1
		WHERE id = $2
	`, postgres.QuoteIdentifier(schemaName))
	_, err = r.db.ExecContext(ctx, query, errMsg, jobID)
	return err
}

func (r *signingRepo) UpdateJobCompleted(ctx context.Context, jobID string, tenantID string, versionID string, signedIPAKey string, manifestKey string) error {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.signing_jobs
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1
	`, schema), jobID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s.application_versions
		SET signing_status = 'signed',
			signed_ipa_s3_key = NULLIF($1, ''),
			manifest_s3_key = NULLIF($2, ''),
			published_at = COALESCE(published_at, NOW())
		WHERE id = $3
	`, schema), signedIPAKey, manifestKey, versionID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *signingRepo) GetJobStatus(ctx context.Context, tenantID string, jobID string) (*domain.SigningJobStatusDTO, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT
			status::text AS status,
			CASE
				WHEN status = 'queued' THEN 10
				WHEN status = 'processing' THEN 50
				WHEN status = 'completed' THEN 100
				WHEN status = 'failed' THEN 100
				ELSE 0
			END AS progress_pct,
			COALESCE(error_message, '') AS error_message
		FROM %s.signing_jobs
		WHERE id = $1
	`, postgres.QuoteIdentifier(schemaName))

	var status domain.SigningJobStatusDTO
	if err := r.db.GetContext(ctx, &status, query, jobID); err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *signingRepo) tenantSchema(ctx context.Context, tenantID string) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}
