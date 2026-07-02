package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

type notificationRepo struct {
	db *postgres.DB
}

func NewNotificationRepository(db *postgres.DB) *notificationRepo {
	return &notificationRepo{db: db}
}

// ListForUser returns the notification feed for a user, including each item's read state.
func (r *notificationRepo) ListForUser(ctx context.Context, tenantID uuid.UUID, userID string) ([]domain.NotificationDTO, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	query := fmt.Sprintf(`
		SELECT
			n.id::text AS id,
			n.title,
			n.body,
			COALESCE(n.target_app_id::text, '') AS target_app_id,
			n.sent_at::text AS sent_at,
			COALESCE(rd.read_at::text, '') AS read_at
		FROM %s.notifications n
		LEFT JOIN %s.user_notification_reads rd
			ON rd.notification_id = n.id AND rd.user_id = NULLIF($1, '')::uuid
		ORDER BY n.sent_at DESC
		LIMIT 100
	`, schema, schema)

	var items []domain.NotificationDTO
	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.NotificationDTO{}
	}
	return items, nil
}

// Create inserts a notification (admin broadcast) and returns its id.
func (r *notificationRepo) Create(ctx context.Context, tenantID uuid.UUID, title, body string, targetAppID *string, nType string) (string, error) {
	if nType == "" {
		nType = "alert"
	}
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf(`
		INSERT INTO %[1]s.notifications (title, body, target_app_id, type)
		VALUES ($1, $2, NULLIF($3, '')::uuid, $4::%[1]s.notification_type)
		RETURNING id::text
	`, postgres.QuoteIdentifier(schemaName))

	var id string
	if err := r.db.QueryRowxContext(ctx, query, title, body, derefOrEmpty(targetAppID), nType).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (r *notificationRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}
