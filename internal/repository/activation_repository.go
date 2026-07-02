package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"platform/internal/store/postgres"
)

var errActivationInvalid = errors.New("invalid activation code")

type activationRepo struct {
	db *postgres.DB
}

func NewActivationRepository(db *postgres.DB) *activationRepo {
	return &activationRepo{db: db}
}

func (r *activationRepo) GenerateCodes(ctx context.Context, tenantID uuid.UUID, count int, codeType, deviceType string, maxDevices, maxUses int, adminID string) ([]string, error) {
	if count <= 0 || count > 500 {
		return nil, errors.New("count must be between 1 and 500")
	}
	if codeType == "" {
		codeType = "usage_count"
	}
	if deviceType == "" {
		deviceType = "both"
	}
	if maxDevices <= 0 {
		maxDevices = 1
	}
	if maxUses <= 0 {
		maxUses = 1
	}

	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, count)
	for len(codes) < count {
		code, err := generateActivationCode()
		if err != nil {
			return nil, err
		}

		query := fmt.Sprintf(`
			INSERT INTO %[1]s.activation_codes (code, type, device_type, max_devices, max_uses, created_by_admin_id)
			VALUES ($1, $2::%[1]s.activation_code_type, $3::%[1]s.allowed_device_type, $4, $5, NULLIF($6, '')::uuid)
			ON CONFLICT (code) DO NOTHING
		`, postgres.QuoteIdentifier(schemaName))

		result, err := r.db.ExecContext(ctx, query, code, codeType, deviceType, maxDevices, maxUses, adminID)
		if err != nil {
			return nil, err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if rowsAffected == 1 {
			codes = append(codes, code)
		}
	}

	return codes, nil
}

func (r *activationRepo) ValidateAndConsume(ctx context.Context, tenantID uuid.UUID, code, deviceID, hardwareType string) (uuid.UUID, error) {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(deviceID) == "" {
		return uuid.Nil, errActivationInvalid
	}

	deviceType := normalizeDeviceType(hardwareType)
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return uuid.Nil, err
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	if err := r.db.SetDedicatedSchema(ctx, tx, schemaName); err != nil {
		return uuid.Nil, err
	}

	var activation struct {
		ID                 uuid.UUID `db:"id"`
		MaxDevices         int       `db:"max_devices"`
		CurrentDeviceCount int       `db:"current_device_count"`
		MaxUses            int       `db:"max_uses"`
		CurrentUses        int       `db:"current_uses"`
	}

	err = tx.GetContext(ctx, &activation, `
		SELECT id, max_devices, current_device_count, max_uses, current_uses
		FROM activation_codes
		WHERE code = $1
			AND is_revoked = false
			AND (expires_at IS NULL OR expires_at > NOW())
			AND (device_type = 'both' OR device_type::text = $2)
		FOR UPDATE
	`, strings.TrimSpace(code), deviceType)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, errActivationInvalid
		}
		return uuid.Nil, err
	}

	if activation.CurrentUses >= activation.MaxUses || activation.CurrentDeviceCount >= activation.MaxDevices {
		return uuid.Nil, errActivationInvalid
	}

	var userID uuid.UUID
	err = tx.GetContext(ctx, &userID, `
		SELECT id
		FROM users
		WHERE activation_code_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, activation.ID)
	if err == sql.ErrNoRows {
		err = tx.QueryRowxContext(ctx, `
			INSERT INTO users (activation_code_id, display_name, last_seen_at)
			VALUES ($1, $2, NOW())
			RETURNING id
		`, activation.ID, "Activated User").Scan(&userID)
	}
	if err != nil {
		return uuid.Nil, err
	}

	var existingDeviceID uuid.UUID
	err = tx.GetContext(ctx, &existingDeviceID, `
		SELECT id
		FROM devices
		WHERE fingerprint_hash = $1
	`, deviceID)
	if err == sql.ErrNoRows {
		err = tx.QueryRowxContext(ctx, `
			INSERT INTO devices (user_id, enrollment_method, fingerprint_hash, device_type, last_seen_at, last_validated_at)
			VALUES ($1, 'fingerprint', $2, $3::device_platform_type, NOW(), NOW())
			RETURNING id
		`, userID, deviceID, deviceType).Scan(&existingDeviceID)
		if err != nil {
			return uuid.Nil, err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE activation_codes
			SET current_device_count = current_device_count + 1
			WHERE id = $1
		`, activation.ID); err != nil {
			return uuid.Nil, err
		}
	} else if err != nil {
		return uuid.Nil, err
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE devices
			SET user_id = $1, is_revoked = false, revoked_at = NULL, last_seen_at = NOW(), last_validated_at = NOW()
			WHERE id = $2
		`, userID, existingDeviceID); err != nil {
			return uuid.Nil, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE activation_codes
		SET current_uses = current_uses + 1,
			first_used_at = COALESCE(first_used_at, NOW())
		WHERE id = $1
	`, activation.ID); err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *activationRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		return "", err
	}
	return schemaName, nil
}

func generateActivationCode() (string, error) {
	var randomBytes [10]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes[:])
	return encoded[:4] + "-" + encoded[4:8] + "-" + encoded[8:12], nil
}

func normalizeDeviceType(deviceType string) string {
	switch strings.ToLower(strings.TrimSpace(deviceType)) {
	case "ipad":
		return "ipad"
	default:
		return "iphone"
	}
}
