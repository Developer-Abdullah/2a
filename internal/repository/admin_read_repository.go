package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

// adminReadRepo provides read-only aggregate queries for the admin dashboard pages.
type adminReadRepo struct {
	db *postgres.DB
}

func NewAdminReadRepository(db *postgres.DB) *adminReadRepo {
	return &adminReadRepo{db: db}
}

func (r *adminReadRepo) Stats(ctx context.Context, tenantID uuid.UUID) (*domain.AdminStatsDTO, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(s)
	query := fmt.Sprintf(`
		SELECT
			(SELECT count(*) FROM %[1]s.applications)::int AS apps,
			(SELECT count(*) FROM %[1]s.application_versions)::int AS versions,
			(SELECT count(*) FROM %[1]s.users)::int AS users,
			(SELECT count(*) FROM %[1]s.devices WHERE is_revoked = false)::int AS devices,
			(SELECT count(*) FROM %[1]s.activation_codes)::int AS codes,
			(SELECT count(*) FROM %[1]s.notifications)::int AS notifications,
			(SELECT COALESCE(avg(rating), 0) FROM %[1]s.platform_ratings)::float8 AS avg_rating,
			(SELECT count(*) FROM %[1]s.platform_ratings)::int AS rating_count
	`, q)

	var stats domain.AdminStatsDTO
	if err := r.db.GetContext(ctx, &stats, query); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *adminReadRepo) ListNotifications(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminNotificationDTO, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT id::text AS id, title, body, type::text AS type, sent_at::text AS sent_at
		FROM %s.notifications ORDER BY sent_at DESC LIMIT 100
	`, postgres.QuoteIdentifier(s))

	var items []domain.AdminNotificationDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminNotificationDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) ListCodes(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminCodeDTO, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT id::text AS id, code, type::text AS type, device_type::text AS device_type,
			max_devices, current_device_count, max_uses, current_uses, is_revoked
		FROM %s.activation_codes ORDER BY first_used_at DESC NULLS LAST, code ASC LIMIT 200
	`, postgres.QuoteIdentifier(s))

	var items []domain.AdminCodeDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminCodeDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminUserDTO, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT u.id::text AS id, COALESCE(u.display_name, '') AS display_name,
			(SELECT count(*) FROM %[1]s.devices d WHERE d.user_id = u.id)::int AS device_count,
			u.created_at::text AS created_at,
			COALESCE(u.last_seen_at::text, '') AS last_seen_at
		FROM %[1]s.users u ORDER BY u.created_at DESC LIMIT 200
	`, postgres.QuoteIdentifier(s))

	var items []domain.AdminUserDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminUserDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) ListDevices(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminDeviceDTO, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT id::text AS id, user_id::text AS user_id, device_type::text AS device_type,
			COALESCE(model_string, '') AS model, enrollment_method::text AS enrollment_method,
			is_revoked, COALESCE(last_seen_at::text, '') AS last_seen_at
		FROM %s.devices ORDER BY last_seen_at DESC NULLS LAST LIMIT 200
	`, postgres.QuoteIdentifier(s))

	var items []domain.AdminDeviceDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminDeviceDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) ListSigningJobs(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminSigningJobDTO, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(s)
	query := fmt.Sprintf(`
		SELECT j.id::text AS id, a.name AS app_name, v.version AS version,
			j.status::text AS status, j.queued_at::text AS created_at
		FROM %[1]s.signing_jobs j
		JOIN %[1]s.application_versions v ON v.id = j.version_id
		JOIN %[1]s.applications a ON a.id = v.application_id
		ORDER BY j.queued_at DESC LIMIT 100
	`, q)
	var items []domain.AdminSigningJobDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminSigningJobDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) ListCertificates(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminCertDTO, error) {
	query := `
		SELECT id::text AS id, label, is_active, added_at::text AS added_at, expires_at::text AS expires_at
		FROM public.tenant_certificates WHERE tenant_id = $1 ORDER BY added_at DESC
	`
	var items []domain.AdminCertDTO
	if err := r.db.SelectContext(ctx, &items, query, tenantID); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminCertDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) ListAudit(ctx context.Context, tenantID uuid.UUID) ([]domain.AdminAuditDTO, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT id::text AS id, action, resource_type, resource_id,
			COALESCE(ip_address::text, '') AS ip_address, created_at::text AS created_at
		FROM %s.audit_logs ORDER BY created_at DESC LIMIT 100
	`, postgres.QuoteIdentifier(s))
	var items []domain.AdminAuditDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminAuditDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) ListAdmins(ctx context.Context) ([]domain.AdminAccountDTO, error) {
	query := `
		SELECT id::text AS id, email, role::text AS role, created_at::text AS created_at
		FROM public.platform_admins ORDER BY created_at ASC
	`
	var items []domain.AdminAccountDTO
	if err := r.db.SelectContext(ctx, &items, query); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.AdminAccountDTO{}
	}
	return items, nil
}

func (r *adminReadRepo) UserDetail(ctx context.Context, tenantID uuid.UUID, userID string) (*domain.AdminUserDetail, error) {
	s, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	q := postgres.QuoteIdentifier(s)

	var user domain.AdminUserDTO
	if err := r.db.GetContext(ctx, &user, fmt.Sprintf(`
		SELECT u.id::text AS id, COALESCE(u.display_name, '') AS display_name,
			(SELECT count(*) FROM %[1]s.devices d WHERE d.user_id = u.id)::int AS device_count,
			u.created_at::text AS created_at, COALESCE(u.last_seen_at::text, '') AS last_seen_at
		FROM %[1]s.users u WHERE u.id = $1::uuid
	`, q), userID); err != nil {
		return nil, err
	}

	var devices []domain.AdminDeviceDTO
	if err := r.db.SelectContext(ctx, &devices, fmt.Sprintf(`
		SELECT id::text AS id, user_id::text AS user_id, device_type::text AS device_type,
			COALESCE(model_string, '') AS model, enrollment_method::text AS enrollment_method,
			is_revoked, COALESCE(last_seen_at::text, '') AS last_seen_at
		FROM %s.devices WHERE user_id = $1::uuid ORDER BY last_seen_at DESC NULLS LAST
	`, q), userID); err != nil {
		return nil, err
	}
	if devices == nil {
		devices = []domain.AdminDeviceDTO{}
	}

	return &domain.AdminUserDetail{User: user, Devices: devices}, nil
}

func (r *adminReadRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}
