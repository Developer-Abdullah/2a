package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

type userRepo struct {
	db *postgres.DB
}

func NewUserRepository(db *postgres.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) GetUserProfile(ctx context.Context, tenantID uuid.UUID, userID string) (*domain.User, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, COALESCE(display_name, '') AS display_name
		FROM %s.users
		WHERE id = $1
	`, postgres.QuoteIdentifier(schemaName))

	var user domain.User
	if err := r.db.GetContext(ctx, &user, query, userID); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetBoundDevices(ctx context.Context, tenantID uuid.UUID, userID string) ([]domain.Device, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, device_type::text AS device_type, is_revoked
		FROM %s.devices
		WHERE user_id = $1
		ORDER BY last_seen_at DESC NULLS LAST
	`, postgres.QuoteIdentifier(schemaName))

	var devices []domain.Device
	if err := r.db.SelectContext(ctx, &devices, query, userID); err != nil {
		return nil, err
	}
	if devices == nil {
		devices = []domain.Device{}
	}
	return devices, nil
}

func (r *userRepo) ReleaseDeviceSlot(ctx context.Context, tenantID uuid.UUID, userID string, deviceID string) error {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE %s.devices
		SET is_revoked = true, revoked_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_revoked = false
	`, postgres.QuoteIdentifier(schemaName))

	result, err := r.db.ExecContext(ctx, query, deviceID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		return "", err
	}
	return schemaName, nil
}
