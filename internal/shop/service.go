// Package shop is the storefront domain: a public catalog of sellable products, customer orders, and
// the checkout -> payment -> fulfillment flow. It reuses the billing provider abstraction (Paymob for
// EGP, MyFatoorah for KWD) to take payment, and mints an activation code per purchased unit once the
// provider confirms the payment via webhook.
package shop

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"platform/internal/billing"
	"platform/internal/domain"
)

// ErrEmptyCart is returned when a checkout has no line items.
var ErrEmptyCart = errors.New("cart is empty")

// ErrUnsupportedCurrency is returned when no payment provider is mapped to the order currency.
var ErrUnsupportedCurrency = errors.New("unsupported currency")

// Repository is the persistence surface the shop service needs, implemented in internal/repository.
type Repository interface {
	ListPublishedProducts(ctx context.Context, tenantID uuid.UUID, currency string) ([]domain.Product, error)
	GetProductBySlug(ctx context.Context, tenantID uuid.UUID, slug, currency string) (*domain.Product, error)
	// CreateOrder prices the cart server-side from the catalog for `currency`, then inserts the order
	// and its items in one transaction, returning the pending order.
	CreateOrder(ctx context.Context, tenantID uuid.UUID, email, phone, currency string, lines []domain.CartLine) (*domain.Order, error)
	GetOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error)
	// AttachCheckout records the chosen provider and its reference on a pending order.
	AttachCheckout(ctx context.Context, tenantID, orderID uuid.UUID, provider, providerRef string) error
	// FulfillByProviderRef idempotently marks the matching order paid+fulfilled and mints an
	// activation code per purchased unit. applied=false means the order was already fulfilled (a
	// replayed webhook), so nothing changed.
	FulfillByProviderRef(ctx context.Context, tenantID uuid.UUID, provider, providerRef string) (order *domain.Order, applied bool, err error)
	// FulfillByID idempotently fulfills a specific order by id — used for manual admin confirmation.
	// applied=false means the order was already fulfilled.
	FulfillByID(ctx context.Context, tenantID, orderID uuid.UUID) (order *domain.Order, applied bool, err error)
	// SubmitProductRating records a 1-5 rating for a product (by slug); returns sql.ErrNoRows if the
	// product does not exist / is unpublished.
	SubmitProductRating(ctx context.Context, tenantID uuid.UUID, slug string, userID *string, rating int, comment *string, ipHash string) error
	ListProductReviews(ctx context.Context, tenantID uuid.UUID, slug string, limit int) ([]domain.ProductReview, error)
	ProductImageKey(ctx context.Context, tenantID uuid.UUID, slug string) (string, error)
	ListOrdersByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]domain.OrderSummary, error)
	// SetOrderProof / OrderProofKey manage the customer's uploaded transfer-screenshot key.
	SetOrderProof(ctx context.Context, tenantID, orderID uuid.UUID, key string) error
	OrderProofKey(ctx context.Context, tenantID, orderID uuid.UUID) (string, error)
}

// FulfillmentNotifier is notified after an order is fulfilled (e.g. to email the codes). It is
// optional and best-effort: an error is logged, never propagated to the payment webhook.
type FulfillmentNotifier interface {
	NotifyFulfilled(ctx context.Context, order *domain.Order) error
}

// Service is the storefront application service.
type Service struct {
	repo     Repository
	registry *billing.Registry
	notifier FulfillmentNotifier
}

func NewService(repo Repository, registry *billing.Registry) *Service {
	return &Service{repo: repo, registry: registry}
}

// SetNotifier attaches an optional post-fulfillment notifier (e.g. code-delivery email).
func (s *Service) SetNotifier(n FulfillmentNotifier) { s.notifier = n }

// providerForCurrency maps a currency to the payment provider that settles it. EGP -> Paymob,
// KWD -> MyFatoorah. Extend here as new markets/providers are added.
func providerForCurrency(currency string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "EGP":
		return "paymob", true
	case "KWD":
		return "myfatoorah", true
	default:
		return "", false
	}
}

func (s *Service) ListProducts(ctx context.Context, tenantID uuid.UUID, currency string) ([]domain.Product, error) {
	return s.repo.ListPublishedProducts(ctx, tenantID, normalizeCurrency(currency))
}

func (s *Service) GetProduct(ctx context.Context, tenantID uuid.UUID, slug, currency string) (*domain.Product, error) {
	return s.repo.GetProductBySlug(ctx, tenantID, slug, normalizeCurrency(currency))
}

func (s *Service) GetOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error) {
	return s.repo.GetOrder(ctx, tenantID, orderID)
}

func (s *Service) SubmitRating(ctx context.Context, tenantID uuid.UUID, slug string, userID *string, rating int, comment *string, ipHash string) error {
	return s.repo.SubmitProductRating(ctx, tenantID, slug, userID, rating, comment, ipHash)
}

func (s *Service) ListReviews(ctx context.Context, tenantID uuid.UUID, slug string, limit int) ([]domain.ProductReview, error) {
	return s.repo.ListProductReviews(ctx, tenantID, slug, limit)
}

func (s *Service) ProductImageKey(ctx context.Context, tenantID uuid.UUID, slug string) (string, error) {
	return s.repo.ProductImageKey(ctx, tenantID, slug)
}

func (s *Service) OrdersByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]domain.OrderSummary, error) {
	return s.repo.ListOrdersByEmail(ctx, tenantID, strings.TrimSpace(email))
}

func (s *Service) SetOrderProof(ctx context.Context, tenantID, orderID uuid.UUID, key string) error {
	return s.repo.SetOrderProof(ctx, tenantID, orderID, key)
}

func (s *Service) OrderProofKey(ctx context.Context, tenantID, orderID uuid.UUID) (string, error) {
	return s.repo.OrderProofKey(ctx, tenantID, orderID)
}

// ConfirmOrder manually fulfills an order (admin-confirmed payment): it mints the codes and, on a
// fresh fulfillment, fires the notifier (code-delivery email). Idempotent — confirming an
// already-fulfilled order is a no-op that still returns the order with its codes.
func (s *Service) ConfirmOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*domain.Order, error) {
	order, applied, err := s.repo.FulfillByID(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	if applied && s.notifier != nil && order != nil {
		if nerr := s.notifier.NotifyFulfilled(ctx, order); nerr != nil {
			log.Ctx(ctx).Error().Err(nerr).Str("order_id", order.ID.String()).Msg("failed to queue confirmation email")
		}
	}
	return order, nil
}

// CreateOrder validates the cart and currency, then persists a pending order priced from the catalog.
func (s *Service) CreateOrder(ctx context.Context, tenantID uuid.UUID, email, phone, currency string, lines []domain.CartLine) (*domain.Order, error) {
	if len(lines) == 0 {
		return nil, ErrEmptyCart
	}
	currency = normalizeCurrency(currency)
	if _, ok := providerForCurrency(currency); !ok {
		return nil, ErrUnsupportedCurrency
	}
	return s.repo.CreateOrder(ctx, tenantID, strings.TrimSpace(email), strings.TrimSpace(phone), currency, lines)
}

// StartCheckout opens a hosted payment for a pending order and returns the redirect URL. The chosen
// provider is derived from the order currency; its reference is stored so the later webhook can be
// correlated back. callbackURL is where the provider returns the buyer (e.g. the order status page).
func (s *Service) StartCheckout(ctx context.Context, tenantID, orderID uuid.UUID, callbackURL string) (string, error) {
	order, err := s.repo.GetOrder(ctx, tenantID, orderID)
	if err != nil {
		return "", err
	}
	if order.Status != "pending" {
		return "", fmt.Errorf("order %s is not payable (status %s)", orderID, order.Status)
	}

	providerName, ok := providerForCurrency(order.Currency)
	if !ok {
		return "", ErrUnsupportedCurrency
	}
	provider, ok := s.registry.Get(providerName)
	if !ok {
		return "", fmt.Errorf("payment provider %q is not configured", providerName)
	}

	result, err := provider.CreateCheckout(ctx, billing.CheckoutRequest{
		TenantID:      order.ID.String(),
		TenantSlug:    displayName(order.Email),
		CustomerEmail: order.Email,
		Amount:        order.Total,
		Currency:      order.Currency,
		CallbackURL:   callbackURL,
	})
	if err != nil {
		return "", err
	}

	if err := s.repo.AttachCheckout(ctx, tenantID, order.ID, provider.Name(), result.ProviderRef); err != nil {
		return "", err
	}
	return result.RedirectURL, nil
}

// HandleWebhook verifies a provider webhook and, on a confirmed payment, fulfills the order. It is
// safe to call repeatedly for the same payment: fulfillment is idempotent on (provider, ref).
func (s *Service) HandleWebhook(ctx context.Context, tenantID uuid.UUID, providerName string, headers http.Header, body []byte) error {
	provider, ok := s.registry.Get(providerName)
	if !ok {
		return fmt.Errorf("unknown payment provider %q", providerName)
	}
	event, err := provider.ParseAndVerifyWebhook(headers, body)
	if err != nil {
		return err
	}
	if !event.Paid {
		return nil // not a successful payment; nothing to fulfill
	}
	order, applied, err := s.repo.FulfillByProviderRef(ctx, tenantID, provider.Name(), event.ProviderRef)
	if err != nil {
		return err
	}
	// Only notify on a fresh fulfillment (applied), so a replayed webhook does not re-send codes.
	// Delivery is best-effort: never fail the webhook if the email cannot be queued.
	if applied && s.notifier != nil && order != nil {
		if nerr := s.notifier.NotifyFulfilled(ctx, order); nerr != nil {
			log.Ctx(ctx).Error().Err(nerr).Str("order_id", order.ID.String()).Msg("failed to queue fulfillment email")
		}
	}
	return nil
}

func normalizeCurrency(currency string) string {
	c := strings.ToUpper(strings.TrimSpace(currency))
	if c == "" {
		return "EGP"
	}
	return c
}

// displayName derives a human name for the provider's customer field from the email local part.
func displayName(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	if email == "" {
		return "customer"
	}
	return email
}
