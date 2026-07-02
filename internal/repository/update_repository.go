package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

type updateRepo struct {
	db *postgres.DB
}

func NewUpdateRepository(db *postgres.DB) *updateRepo {
	return &updateRepo{db: db}
}

// ListUpdates returns published update transitions (from -> to version) for the tenant.
func (r *updateRepo) ListUpdates(ctx context.Context, tenantID uuid.UUID) ([]domain.UpdateDTO, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	schema := postgres.QuoteIdentifier(schemaName)

	query := fmt.Sprintf(`
		SELECT
			u.id::text AS id,
			u.application_id::text AS app_id,
			fv.version AS current_version,
			tv.version AS latest_version,
			u.update_type::text AS update_type
		FROM %s.updates u
		JOIN %s.application_versions fv ON fv.id = u.from_version_id
		JOIN %s.application_versions tv ON tv.id = u.to_version_id
		ORDER BY u.published_at DESC
	`, schema, schema, schema)

	var updates []domain.UpdateDTO
	if err := r.db.SelectContext(ctx, &updates, query); err != nil {
		return nil, err
	}
	if updates == nil {
		updates = []domain.UpdateDTO{}
	}
	return updates, nil
}

func (r *updateRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}
