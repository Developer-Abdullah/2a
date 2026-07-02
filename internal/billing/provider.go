// Package billing implements subscription checkout and provider webhooks behind a provider-agnostic
// interface, so the platform can run MyFatoorah (KWD/GCC) and Paymob (EGP/Egypt) side by side.
//
// Security posture: webhook handlers NEVER trust an unverified payload. Each provider verifies its
// own signature (HMAC) before the event is applied, and the service layer dedupes by idempotency
// key so a replayed or double-delivered webhook cannot credit a subscription twice.
package billing

import (
	"context"
	"errors"
	"net/http"
)

// CheckoutRequest is what the app hands a provider to start a hosted checkout.
type CheckoutRequest struct {
	TenantID      string
	TenantSlug    string
	PlanID        string
	CustomerEmail string
	Amount        float64 // in the plan's currency major units (e.g. 5.000 KWD)
	Currency      string
	CallbackURL   string // where the provider redirects the user after payment
}

// CheckoutResult is what the provider returns: where to send the user, and the reference we store
// so the later webhook can be correlated back to this checkout.
type CheckoutResult struct {
	RedirectURL string
	ProviderRef string
}

// WebhookEvent is the normalized, signature-verified outcome of a provider callback.
type WebhookEvent struct {
	ProviderRef string
	Paid        bool
	Amount      float64
	Currency    string
	// IdempotencyKey uniquely identifies this payment across redeliveries (provider + ref).
	IdempotencyKey string
	Raw           []byte
}

// Provider is a payment gateway adapter.
type Provider interface {
	Name() string
	CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error)
	// ParseAndVerifyWebhook verifies the signature and returns the normalized event. It MUST return
	// an error (not a zero event) when the signature is missing or invalid.
	ParseAndVerifyWebhook(headers http.Header, body []byte) (*WebhookEvent, error)
}

// ErrInvalidSignature is returned when a webhook fails signature verification.
var ErrInvalidSignature = errors.New("invalid webhook signature")

// Registry holds the configured providers, keyed by their Name().
type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers ...Provider) *Registry {
	m := map[string]Provider{}
	for _, p := range providers {
		if p != nil {
			m[p.Name()] = p
		}
	}
	return &Registry{providers: m}
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.providers))
	for name := range r.providers {
		out = append(out, name)
	}
	return out
}
