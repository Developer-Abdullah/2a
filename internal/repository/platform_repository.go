package repository

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"platform/internal/domain"
	"platform/internal/provision"
	"platform/internal/store/postgres"
)

type platformRepo struct {
	db *postgres.DB
}

func NewPlatformRepository(db *postgres.DB) *platformRepo {
	return &platformRepo{db: db}
}

// ListTenants returns every merchant with a best-effort app count from its own schema.
func (r *platformRepo) ListTenants(ctx context.Context) ([]domain.TenantSummaryDTO, error) {
	query := `
		SELECT t.slug, COALESCE(c.app_name, t.slug) AS name, t.schema_name,
			t.status::text AS status, t.created_at::text AS created_at
		FROM public.tenants t
		LEFT JOIN public.tenant_config c ON c.tenant_id = t.id
		ORDER BY t.created_at ASC
	`
	var items []domain.TenantSummaryDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	for i := range items {
		var n int
		if err := r.db.GetContext(ctx, &n, fmt.Sprintf(
			"SELECT count(*) FROM %s.applications", postgres.QuoteIdentifier(items[i].SchemaName))); err == nil {
			items[i].Apps = n
		}
	}
	if items == nil {
		items = []domain.TenantSummaryDTO{}
	}
	return items, nil
}

// CreateTenant onboards a merchant end-to-end: public rows (tenant, config, admin, cert) plus a
// freshly provisioned, fully migrated schema.
func (r *platformRepo) CreateTenant(ctx context.Context, slug, name, adminEmail, adminPassword string) error {
	schema := "tenant_" + strings.ReplaceAll(slug, "-", "_")

	var tenantID string
	if err := r.db.QueryRowxContext(ctx, `
		INSERT INTO public.tenants (slug, plan_id, isolation_mode, schema_name, s3_prefix)
		VALUES ($1, (SELECT id FROM public.tenant_plans LIMIT 1), 'dedicated_schema', $2, $3)
		RETURNING id::text
	`, slug, schema, schema+"/").Scan(&tenantID); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO public.tenant_config (tenant_id, app_name, support_email)
		VALUES ($1::uuid, $2, $3)
		ON CONFLICT (tenant_id) DO UPDATE SET app_name = EXCLUDED.app_name
	`, tenantID, name, "support@"+slug+".local"); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO public.platform_admins (email, password_hash, role, tenant_id)
		VALUES ($1, $2, 'super_admin', $3::uuid)
		ON CONFLICT (email) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
	`, adminEmail, string(hash), tenantID); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO public.tenant_certificates (tenant_id, label, s3_key_p12_enc, s3_key_mobileprovision, encrypted_password, expires_at)
		VALUES ($1::uuid, $2, 'demo/cert.p12.enc', 'demo/profile.mobileprovision', $3, NOW() + INTERVAL '365 days')
	`, tenantID, name+" Cert", []byte("demo-encrypted-password")); err != nil {
		return err
	}

	return provision.ProvisionTenantSchema(ctx, r.db, schema)
}

func (r *platformRepo) SetTenantStatus(ctx context.Context, slug, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE public.tenants SET status = $2::tenant_status WHERE slug = $1`, slug, status)
	return err
}
