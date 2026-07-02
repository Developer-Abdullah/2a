package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"platform/internal/domain"
	"platform/internal/store/postgres"
)

type ratingRepo struct {
	db *postgres.DB
}

func NewRatingRepository(db *postgres.DB) *ratingRepo {
	return &ratingRepo{db: db}
}

// SubmitRating records a platform rating. The (ip_hash, rating_date) unique index enforces
// one rating per IP per day; a repeat from the same IP that day updates the existing row.
func (r *ratingRepo) SubmitRating(ctx context.Context, tenantID uuid.UUID, userID, deviceID *string, rating int, comment *string, ipHash string) error {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.platform_ratings (user_id, device_id, rating, comment, ip_hash)
		VALUES (NULLIF($1, '')::uuid, NULLIF($2, '')::uuid, $3, NULLIF($4, ''), $5)
		ON CONFLICT (ip_hash, rating_date)
		DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment, updated_at = NOW()
	`, postgres.QuoteIdentifier(schemaName))

	_, err = r.db.ExecContext(ctx, query, derefOrEmpty(userID), derefOrEmpty(deviceID), rating, derefOrEmpty(comment), ipHash)
	return err
}

// GetSummary returns the average, total count, and per-star breakdown of all ratings.
func (r *ratingRepo) GetSummary(ctx context.Context, tenantID uuid.UUID) (*domain.RatingSummary, error) {
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(AVG(rating), 0)::float8 AS average,
			COUNT(*)::int AS count,
			COALESCE(SUM((rating = 1)::int), 0)::int AS r1,
			COALESCE(SUM((rating = 2)::int), 0)::int AS r2,
			COALESCE(SUM((rating = 3)::int), 0)::int AS r3,
			COALESCE(SUM((rating = 4)::int), 0)::int AS r4,
			COALESCE(SUM((rating = 5)::int), 0)::int AS r5
		FROM %s.platform_ratings
	`, postgres.QuoteIdentifier(schemaName))

	var summary domain.RatingSummary
	if err := r.db.GetContext(ctx, &summary, query); err != nil {
		return nil, err
	}
	return &summary, nil
}

// ListRecent returns the latest ratings for the admin dashboard.
func (r *ratingRepo) ListRecent(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.PlatformRating, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	schemaName, err := r.tenantSchema(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, rating, comment, ip_hash, device_id, created_at, updated_at
		FROM %s.platform_ratings
		ORDER BY created_at DESC
		LIMIT $1
	`, postgres.QuoteIdentifier(schemaName))

	var ratings []domain.PlatformRating
	if err := r.db.SelectContext(ctx, &ratings, query, limit); err != nil {
		return nil, err
	}
	if ratings == nil {
		ratings = []domain.PlatformRating{}
	}
	return ratings, nil
}

func (r *ratingRepo) tenantSchema(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var schemaName string
	if err := r.db.GetContext(ctx, &schemaName, `SELECT schema_name FROM public.tenants WHERE id = $1`, tenantID); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	return schemaName, nil
}

// derefOrEmpty turns an optional string pointer into a plain string ("" when nil).
// Shared across repositories in this package.
func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
