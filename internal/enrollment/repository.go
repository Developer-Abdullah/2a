package enrollment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"platform/internal/store/postgres"
)

type Session struct {
	ID           uuid.UUID `db:"id"`
	TenantID     uuid.UUID `db:"tenant_id"`
	OneTimeToken string    `db:"one_time_token"`
	IPAddress    string    `db:"ip_address"`
	UDIDHash     string    `db:"udid_hash"`
	DeviceType   string    `db:"device_type"`
	Completed    bool      `db:"completed"`
	CreatedAt    time.Time `db:"created_at"`
	ExpiresAt    time.Time `db:"expires_at"`
}

// SessionRepository persists UDID enrollment sessions inside a tenant's dedicated schema. The
// schema name is passed in explicitly (resolved from the request's tenant) so a single pooled
// connection can serve every tenant without relying on search_path.
type SessionRepository interface {
	CreateSession(ctx context.Context, schema string, session *Session) error
	GetSessionByToken(ctx context.Context, schema, token string) (*Session, error)
	CompleteSession(ctx context.Context, schema, token, udidHash, deviceType string) error
}

type sessionRepo struct {
	db *postgres.DB
}

func NewSessionRepository(db *postgres.DB) SessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) CreateSession(ctx context.Context, schema string, session *Session) error {
	q := fmt.Sprintf(`
		INSERT INTO %s.udid_enrollment_sessions (tenant_id, one_time_token, ip_address, expires_at)
		VALUES ($1, $2, NULLIF($3, '')::inet, $4)
	`, postgres.QuoteIdentifier(schema))
	_, err := r.db.ExecContext(ctx, q, session.TenantID, session.OneTimeToken, session.IPAddress, session.ExpiresAt)
	return err
}

func (r *sessionRepo) GetSessionByToken(ctx context.Context, schema, token string) (*Session, error) {
	q := fmt.Sprintf(`
		SELECT id, tenant_id, one_time_token,
			COALESCE(host(ip_address), '') AS ip_address,
			COALESCE(udid_hash, '') AS udid_hash,
			COALESCE(device_type::text, '') AS device_type,
			completed, created_at, expires_at
		FROM %s.udid_enrollment_sessions
		WHERE one_time_token = $1
	`, postgres.QuoteIdentifier(schema))

	var s Session
	if err := r.db.GetContext(ctx, &s, q, token); err != nil {
		return nil, err
	}
	return &s, nil
}

// CompleteSession records the captured UDID against an unexpired session. A no-op (expired or
// unknown token) is reported so the caller can surface it rather than silently succeeding.
func (r *sessionRepo) CompleteSession(ctx context.Context, schema, token, udidHash, deviceType string) error {
	q := fmt.Sprintf(`
		UPDATE %[1]s.udid_enrollment_sessions
		SET completed = true, udid_hash = $2,
			device_type = NULLIF($3, '')::%[1]s.device_platform_type
		WHERE one_time_token = $1 AND expires_at > NOW()
	`, postgres.QuoteIdentifier(schema))
	res, err := r.db.ExecContext(ctx, q, token, udidHash, deviceType)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("no active enrollment session for token")
	}
	return nil
}
