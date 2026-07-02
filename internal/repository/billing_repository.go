package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"platform/internal/billing"
	"platform/internal/store/postgres"
)

type billingRepo struct {
	db *postgres.DB
}

func NewBillingRepository(db *postgres.DB) *billingRepo {
	return &billingRepo{db: db}
}

func (r *billingRepo) PlanPricing(ctx context.Context, planID string) (*billing.PlanPricing, error) {
	var p billing.PlanPricing
	err := r.db.QueryRowxContext(ctx, `
		SELECT price_amount, currency, billing_interval
		FROM public.tenant_plans WHERE id = $1::uuid
	`, planID).Scan(&p.Amount, &p.Currency, &p.Interval)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *billingRepo) CheckoutContext(ctx context.Context, tenantID string) (*billing.CheckoutContext, error) {
	var c billing.CheckoutContext
	err := r.db.QueryRowxContext(ctx, `
		SELECT t.slug, COALESCE(cfg.support_email, '')
		FROM public.tenants t
		LEFT JOIN public.tenant_config cfg ON cfg.tenant_id = t.id
		WHERE t.id = $1::uuid
	`, tenantID).Scan(&c.Slug, &c.Email)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CreatePendingPayment inserts a pending payment keyed by (provider, provider_ref). A duplicate
// checkout for the same reference is ignored so retries don't create ledger noise.
func (r *billingRepo) CreatePendingPayment(ctx context.Context, tenantID, planID, provider, providerRef string, amount float64, currency string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO public.subscription_payments
			(tenant_id, plan_id, provider, provider_ref, idempotency_key, amount, currency, status)
		VALUES ($1::uuid, $2::uuid, $3, $4, $3 || ':' || $4, $5, $6, 'pending')
		ON CONFLICT (idempotency_key) DO NOTHING
	`, tenantID, planID, provider, providerRef, amount, currency)
	return err
}

// MarkPaidAndExtend applies a successful payment inside one transaction. It locks the payment row,
// no-ops if it is already paid (idempotent under webhook redelivery), otherwise marks it paid and
// extends the tenant's subscription by the plan's interval, resetting the grace window.
func (r *billingRepo) MarkPaidAndExtend(ctx context.Context, provider, providerRef string, graceDays int) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var (
		paymentID string
		tenantID  string
		planID    string
		status    string
	)
	err = tx.QueryRowxContext(ctx, `
		SELECT id::text, tenant_id::text, plan_id::text, status
		FROM public.subscription_payments
		WHERE provider = $1 AND provider_ref = $2
		FOR UPDATE
	`, provider, providerRef).Scan(&paymentID, &tenantID, &planID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("no pending payment for %s ref %s", provider, providerRef)
	}
	if err != nil {
		return false, err
	}
	if status == "paid" {
		return false, nil // already applied — idempotent no-op
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE public.subscription_payments SET status = 'paid', paid_at = NOW() WHERE id = $1::uuid
	`, paymentID); err != nil {
		return false, err
	}

	var interval string
	if err := tx.QueryRowxContext(ctx, `SELECT billing_interval FROM public.tenant_plans WHERE id = $1::uuid`, planID).Scan(&interval); err != nil {
		return false, err
	}
	intervalSQL := "1 year"
	if interval == "monthly" {
		intervalSQL = "1 month"
	}

	// New expiry = max(now, current expiry) + interval, so early renewals stack instead of truncating.
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE public.tenants
		SET plan_id = $2::uuid,
			subscription_status = 'active',
			subscription_expires_at = GREATEST(NOW(), COALESCE(subscription_expires_at, NOW())) + INTERVAL '%s',
			grace_until = GREATEST(NOW(), COALESCE(subscription_expires_at, NOW())) + INTERVAL '%s' + ($3 || ' days')::interval
		WHERE id = $1::uuid
	`, intervalSQL, intervalSQL), tenantID, planID, graceDays); err != nil {
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO public.subscription_events (tenant_id, event_type, detail)
		VALUES ($1::uuid, 'activated', jsonb_build_object('provider', $2, 'provider_ref', $3))
	`, tenantID, provider, providerRef); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (r *billingRepo) Subscription(ctx context.Context, tenantID string) (*billing.SubscriptionSnapshot, error) {
	var (
		status  string
		expires sql.NullTime
		grace   sql.NullTime
	)
	err := r.db.QueryRowxContext(ctx, `
		SELECT subscription_status::text, subscription_expires_at, grace_until
		FROM public.tenants WHERE id = $1::uuid
	`, tenantID).Scan(&status, &expires, &grace)
	if err != nil {
		return nil, err
	}
	snap := &billing.SubscriptionSnapshot{Status: status}
	if expires.Valid {
		snap.ExpiresAt = &expires.Time
	}
	if grace.Valid {
		snap.GraceUntil = &grace.Time
	}
	return snap, nil
}

// ExpiringWithin returns active tenants expiring within `days` that have not already been reminded
// for this exact expiry timestamp (dedupe via the subscription_events log).
func (r *billingRepo) ExpiringWithin(ctx context.Context, days int) ([]billing.ReminderTarget, error) {
	rows, err := r.db.QueryxContext(ctx, `
		SELECT t.id::text, t.subscription_expires_at
		FROM public.tenants t
		WHERE t.subscription_expires_at IS NOT NULL
		  AND t.subscription_expires_at BETWEEN NOW() AND NOW() + ($1 || ' days')::interval
		  AND NOT EXISTS (
			SELECT 1 FROM public.subscription_events e
			WHERE e.tenant_id = t.id AND e.event_type = 'reminder'
			  AND e.detail->>'expires_at' = t.subscription_expires_at::text
		  )
	`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []billing.ReminderTarget
	for rows.Next() {
		var tgt billing.ReminderTarget
		if err := rows.Scan(&tgt.TenantID, &tgt.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, tgt)
	}
	return out, rows.Err()
}

func (r *billingRepo) RecordReminder(ctx context.Context, tenantID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO public.subscription_events (tenant_id, event_type, detail)
		VALUES ($1::uuid, 'reminder', jsonb_build_object('expires_at', $2::text))
	`, tenantID, expiresAt)
	return err
}

// SweepStatuses advances subscription statuses based on the clock: active tenants past their expiry
// but inside grace become past_due; tenants past their grace window become expired.
func (r *billingRepo) SweepStatuses(ctx context.Context, graceDays int) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE public.tenants
		SET subscription_status = 'past_due'
		WHERE subscription_status = 'active'
		  AND subscription_expires_at IS NOT NULL
		  AND NOW() >= subscription_expires_at
		  AND (grace_until IS NULL OR NOW() < grace_until)
	`); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE public.tenants
		SET subscription_status = 'expired'
		WHERE subscription_status IN ('active', 'past_due')
		  AND grace_until IS NOT NULL
		  AND NOW() >= grace_until
	`)
	return err
}
