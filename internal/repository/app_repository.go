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

type appRepo struct {
	db *postgres.DB
}

func NewAppRepository(db *postgres.DB) *appRepo {
	return &appRepo{db: db}
}

func (r *appRepo) ListApps(ctx context.Context, tenantID uuid.UUID, cursor string, limit int, category string) ([]domain.AppItemDTO, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, "", err
	}

	args := []interface{}{limit + 1}
	conditions := []string{"is_published = true"}

	if category != "" {
		args = append(args, category)
		conditions = append(conditions, fmt.Sprintf("category::text = $%d", len(args)))
	}
	if cursor != "" {
		args = append(args, cursor)
		conditions = append(conditions, fmt.Sprintf("created_at < $%d::timestamptz", len(args)))
	}

	query := fmt.Sprintf(`
		SELECT
			id::text AS id,
			name,
			category::text AS category,
			COALESCE(icon_s3_key, '') AS icon_url,
			bundle_identifier,
			is_published,
			created_at::text AS created_at
		FROM %s.applications
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $1
	`, postgres.QuoteIdentifier(schemaName), strings.Join(conditions, " AND "))

	var items []domain.AppItemDTO
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, "", err
	}
	if items == nil {
		items = []domain.AppItemDTO{}
	}

	nextCursor := ""
	if len(items) > limit {
		nextCursor = items[limit-1].CreatedAt
		items = items[:limit]
	}

	return items, nextCursor, nil
}

func (r *appRepo) GetAppByID(ctx context.Context, tenantID uuid.UUID, appID string) (*domain.AppItemDTO, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT
			id::text AS id,
			name,
			category::text AS category,
			COALESCE(icon_s3_key, '') AS icon_url,
			bundle_identifier,
			is_published,
			created_at::text AS created_at
		FROM %s.applications
		WHERE id = $1 AND is_published = true
	`, postgres.QuoteIdentifier(schemaName))

	var app domain.AppItemDTO
	if err := r.db.GetContext(ctx, &app, query, appID); err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}
