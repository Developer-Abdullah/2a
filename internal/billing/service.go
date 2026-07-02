package billing

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// PlanPricing is the billable shape of a plan.
type PlanPricing struct {
	Amount   float64
	Currency string
	Interval string // "monthly" | "yearly"
}

// CheckoutContext is the tenant data needed to open a checkout.
type CheckoutContext struct {
	Slug  string
	Email string
}

// SubscriptionSnapshot is the current billing state of a tenant.
type SubscriptionSnapshot struct {
	Status    string
	ExpiresAt *time.Time
	GraceUntil *time.Time
}

// Repository is the persistence surface the billing service needs. Implemented in internal/repository.
type Repository interface {
	PlanPricing(ctx context.Context, planID string) (*PlanPricing, error)
	CheckoutContext(ctx context.Context, tenantID string) (*CheckoutContext, error)
	// CreatePendingPayment records a checkout keyed by (provider, providerRef); a duplicate is a no-op.
	CreatePendingPayment(ctx context.Context, tenantID, planID, provider, providerRef string, amount float64, currency string) error
	// MarkPaidAndExtend idempotently applies a successful payment: it returns applied=false if the
	// (provider, providerRef) payment was already marked paid, so a replayed webhook is a no-op.
	MarkPaidAndExtend(ctx context.Context, provider, providerRef string, graceDays int) (applied bool, err error)
	Subscription(ctx context.Context, tenantID string) (*SubscriptionSnapshot, error)
	// ExpiringWithin lists tenants whose subscription expires within `days` and that have not yet been
	// reminded for that expiry, returning (tenantID, expiresAt).
	ExpiringWithin(ctx context.Context, days int) ([]ReminderTarget, error)
	RecordReminder(ctx context.Context, tenantID string, expiresAt time.Time) error
	// SweepStatuses transitions tenants to past_due (past expiry, within grace) and expired (past grace).
	SweepStatuses(ctx context.Context, graceDays int) error
}

// ReminderTarget is one tenant flagged for a renewal reminder.
type ReminderTarget struct {
	TenantID  string
	ExpiresAt time.Time
}

type Service struct {
	repo      Repository
	registry  *Registry
	graceDays int
}

func NewService(repo Repository, registry *Registry, graceDays int) *Service {
	if graceDays < 0 {
		graceDays = 0
	}
	return &Service{repo: repo, registry: registry, graceDays: graceDays}
}

func (s *Service) GraceDays() int { return s.graceDays }

// StartCheckout resolves the plan price, opens a provider checkout, records the pending payment, and
// returns the redirect URL for the merchant.
func (s *Service) StartCheckout(ctx context.Context, providerName, tenantID, planID, callbackURL string) (string, error) {
	provider, ok := s.registry.Get(providerName)
	if !ok {
		return "", fmt.Errorf("unknown payment provider %q", providerName)
	}
	pricing, err := s.repo.PlanPricing(ctx, planID)
	if err != nil {
		return "", err
	}
	tctx, err := s.repo.CheckoutContext(ctx, tenantID)
	if err != nil {
		return "", err
	}

	result, err := provider.CreateCheckout(ctx, CheckoutRequest{
		TenantID: tenantID, TenantSlug: tctx.Slug, PlanID: planID,
		CustomerEmail: tctx.Email, Amount: pricing.Amount, Currency: pricing.Currency,
		CallbackURL: callbackURL,
	})
	if err != nil {
		return "", err
	}

	if err := s.repo.CreatePendingPayment(ctx, tenantID, planID, provider.Name(), result.ProviderRef, pricing.Amount, pricing.Currency); err != nil {
		return "", err
	}
	return result.RedirectURL, nil
}

// HandleWebhook verifies a provider webhook and, on a successful payment, extends the subscription.
// It is safe to call repeatedly for the same payment: MarkPaidAndExtend is idempotent.
func (s *Service) HandleWebhook(ctx context.Context, providerName string, headers http.Header, body []byte) error {
	provider, ok := s.registry.Get(providerName)
	if !ok {
		return fmt.Errorf("unknown payment provider %q", providerName)
	}
	event, err := provider.ParseAndVerifyWebhook(headers, body)
	if err != nil {
		return err
	}
	if !event.Paid {
		return nil // not a successful payment; nothing to apply
	}
	_, err = s.repo.MarkPaidAndExtend(ctx, provider.Name(), event.ProviderRef, s.graceDays)
	return err
}

// DistributionAllowed reports whether a tenant may currently publish/sign new builds. Existing
// signed builds are unaffected by this check — it gates new submissions only. A tenant is allowed
// while trialing, while active, or while still inside the grace window after expiry.
func (s *Service) DistributionAllowed(ctx context.Context, tenantID string, now time.Time) (bool, error) {
	snap, err := s.repo.Subscription(ctx, tenantID)
	if err != nil {
		return false, err
	}
	return snapshotAllows(snap, now), nil
}

// SubscriptionView is the subscription snapshot plus the derived distribution decision, for display.
type SubscriptionView struct {
	Status     string
	ExpiresAt  *time.Time
	GraceUntil *time.Time
	Allowed    bool
}

// SubscriptionView returns the tenant's subscription state and whether distribution is currently open.
func (s *Service) SubscriptionView(ctx context.Context, tenantID string) (*SubscriptionView, error) {
	snap, err := s.repo.Subscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return &SubscriptionView{
		Status: snap.Status, ExpiresAt: snap.ExpiresAt, GraceUntil: snap.GraceUntil,
		Allowed: snapshotAllows(snap, time.Now()),
	}, nil
}

// snapshotAllows is the pure grace-period decision, split out for testing.
func snapshotAllows(snap *SubscriptionSnapshot, now time.Time) bool {
	if snap == nil {
		return false
	}
	// Never-subscribed / trial tenants (no expiry recorded yet) are allowed.
	if snap.ExpiresAt == nil {
		return snap.Status != "expired"
	}
	// Within the paid period.
	if now.Before(*snap.ExpiresAt) {
		return true
	}
	// Past expiry: allowed only while inside the grace window.
	if snap.GraceUntil != nil {
		return now.Before(*snap.GraceUntil)
	}
	return false
}

// SendRenewalReminders records a reminder event for each tenant expiring within `days`.
func (s *Service) SendRenewalReminders(ctx context.Context, days int) (int, error) {
	targets, err := s.repo.ExpiringWithin(ctx, days)
	if err != nil {
		return 0, err
	}
	for _, t := range targets {
		if err := s.repo.RecordReminder(ctx, t.TenantID, t.ExpiresAt); err != nil {
			return 0, err
		}
	}
	return len(targets), nil
}

// SweepStatuses advances subscription statuses (active -> past_due -> expired) based on the clock.
func (s *Service) SweepStatuses(ctx context.Context) error {
	return s.repo.SweepStatuses(ctx, s.graceDays)
}
