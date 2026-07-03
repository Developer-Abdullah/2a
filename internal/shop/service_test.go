package shop

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"platform/internal/billing"
	"platform/internal/domain"
)

// fakeProvider is a billing.Provider stub whose checkout ref and webhook outcome are configurable.
type fakeProvider struct {
	name  string
	ref   string
	event *billing.WebhookEvent
	err   error
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) CreateCheckout(_ context.Context, _ billing.CheckoutRequest) (*billing.CheckoutResult, error) {
	return &billing.CheckoutResult{RedirectURL: "https://pay.test/" + f.ref, ProviderRef: f.ref}, nil
}
func (f *fakeProvider) ParseAndVerifyWebhook(http.Header, []byte) (*billing.WebhookEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.event, nil
}

// fakeRepo records the calls the service makes so we can assert on provider selection and fulfillment.
type fakeRepo struct {
	order          *domain.Order
	attachedProv   string
	attachedRef    string
	fulfillCalls   int
	fulfillApplied bool
}

func (r *fakeRepo) ListPublishedProducts(context.Context, uuid.UUID, string) ([]domain.Product, error) {
	return nil, nil
}
func (r *fakeRepo) GetProductBySlug(context.Context, uuid.UUID, string, string) (*domain.Product, error) {
	return nil, nil
}
func (r *fakeRepo) CreateOrder(_ context.Context, _ uuid.UUID, email, phone, currency string, lines []domain.CartLine) (*domain.Order, error) {
	return &domain.Order{ID: uuid.New(), Email: email, Phone: phone, Currency: currency, Status: "pending", Total: 10}, nil
}
func (r *fakeRepo) GetOrder(context.Context, uuid.UUID, uuid.UUID) (*domain.Order, error) {
	return r.order, nil
}
func (r *fakeRepo) AttachCheckout(_ context.Context, _ uuid.UUID, _ uuid.UUID, provider, ref string) error {
	r.attachedProv = provider
	r.attachedRef = ref
	return nil
}
func (r *fakeRepo) FulfillByProviderRef(context.Context, uuid.UUID, string, string) (*domain.Order, bool, error) {
	r.fulfillCalls++
	return r.order, r.fulfillApplied, nil
}
func (r *fakeRepo) FulfillByID(context.Context, uuid.UUID, uuid.UUID) (*domain.Order, bool, error) {
	r.fulfillCalls++
	return r.order, r.fulfillApplied, nil
}
func (r *fakeRepo) SubmitProductRating(context.Context, uuid.UUID, string, *string, int, *string, string) error {
	return nil
}
func (r *fakeRepo) ListProductReviews(context.Context, uuid.UUID, string, int) ([]domain.ProductReview, error) {
	return nil, nil
}
func (r *fakeRepo) ProductImageKey(context.Context, uuid.UUID, string) (string, error) {
	return "", nil
}
func (r *fakeRepo) ListOrdersByEmail(context.Context, uuid.UUID, string) ([]domain.OrderSummary, error) {
	return nil, nil
}
func (r *fakeRepo) SetOrderProof(context.Context, uuid.UUID, uuid.UUID, string) error {
	return nil
}
func (r *fakeRepo) OrderProofKey(context.Context, uuid.UUID, uuid.UUID) (string, error) {
	return "", nil
}
func (r *fakeRepo) CodeStatus(context.Context, uuid.UUID, string) (*domain.CodeStatus, error) {
	return nil, nil
}

func TestProviderForCurrency(t *testing.T) {
	cases := map[string]struct {
		want string
		ok   bool
	}{
		"EGP": {"paymob", true},
		"egp": {"paymob", true},
		"KWD": {"myfatoorah", true},
		"USD": {"", false},
	}
	for cur, tc := range cases {
		got, ok := providerForCurrency(cur)
		if got != tc.want || ok != tc.ok {
			t.Errorf("providerForCurrency(%q) = (%q,%v), want (%q,%v)", cur, got, ok, tc.want, tc.ok)
		}
	}
}

func TestCreateOrder_Validation(t *testing.T) {
	svc := NewService(&fakeRepo{}, billing.NewRegistry())

	if _, err := svc.CreateOrder(context.Background(), uuid.New(), "a@b.com", "", "EGP", nil); !errors.Is(err, ErrEmptyCart) {
		t.Errorf("empty cart: want ErrEmptyCart, got %v", err)
	}
	if _, err := svc.CreateOrder(context.Background(), uuid.New(), "a@b.com", "", "USD", []domain.CartLine{{Slug: "x", Qty: 1}}); !errors.Is(err, ErrUnsupportedCurrency) {
		t.Errorf("bad currency: want ErrUnsupportedCurrency, got %v", err)
	}
}

func TestStartCheckout_SelectsProviderByCurrency(t *testing.T) {
	orderID := uuid.New()
	repo := &fakeRepo{order: &domain.Order{ID: orderID, Currency: "KWD", Status: "pending", Total: 3.294, Email: "a@b.com"}}
	reg := billing.NewRegistry(&fakeProvider{name: "myfatoorah", ref: "INV-1"}, &fakeProvider{name: "paymob", ref: "ORD-1"})
	svc := NewService(repo, reg)

	url, err := svc.StartCheckout(context.Background(), uuid.New(), orderID, "https://store.test/order/x")
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if repo.attachedProv != "myfatoorah" || repo.attachedRef != "INV-1" {
		t.Errorf("KWD should select myfatoorah/INV-1, got %s/%s", repo.attachedProv, repo.attachedRef)
	}
	if url != "https://pay.test/INV-1" {
		t.Errorf("unexpected redirect url %q", url)
	}
}

func TestHandleWebhook_FulfillsOnlyWhenPaid(t *testing.T) {
	// Unpaid event: no fulfillment.
	repo := &fakeRepo{order: &domain.Order{Status: "pending"}}
	reg := billing.NewRegistry(&fakeProvider{name: "paymob", event: &billing.WebhookEvent{Paid: false, ProviderRef: "ORD-9"}})
	svc := NewService(repo, reg)
	if err := svc.HandleWebhook(context.Background(), uuid.New(), "paymob", http.Header{}, nil); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if repo.fulfillCalls != 0 {
		t.Errorf("unpaid event must not fulfill, got %d calls", repo.fulfillCalls)
	}

	// Paid event: fulfillment invoked once.
	repo2 := &fakeRepo{order: &domain.Order{Status: "fulfilled"}, fulfillApplied: true}
	reg2 := billing.NewRegistry(&fakeProvider{name: "paymob", event: &billing.WebhookEvent{Paid: true, ProviderRef: "ORD-9"}})
	svc2 := NewService(repo2, reg2)
	if err := svc2.HandleWebhook(context.Background(), uuid.New(), "paymob", http.Header{}, nil); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if repo2.fulfillCalls != 1 {
		t.Errorf("paid event must fulfill once, got %d calls", repo2.fulfillCalls)
	}
}

func TestConfirmOrder_NotifiesOnFreshFulfillment(t *testing.T) {
	// Fresh fulfillment (applied): notifier is invoked.
	notified := 0
	repo := &fakeRepo{order: &domain.Order{ID: uuid.New(), Status: "fulfilled"}, fulfillApplied: true}
	svc := NewService(repo, billing.NewRegistry())
	svc.SetNotifier(notifierFunc(func(context.Context, *domain.Order) error { notified++; return nil }))
	if _, err := svc.ConfirmOrder(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if repo.fulfillCalls != 1 || notified != 1 {
		t.Errorf("fresh confirm: fulfillCalls=%d notified=%d, want 1/1", repo.fulfillCalls, notified)
	}

	// Already fulfilled (applied=false): no notification (no double-send of codes).
	notified2 := 0
	repo2 := &fakeRepo{order: &domain.Order{ID: uuid.New(), Status: "fulfilled"}, fulfillApplied: false}
	svc2 := NewService(repo2, billing.NewRegistry())
	svc2.SetNotifier(notifierFunc(func(context.Context, *domain.Order) error { notified2++; return nil }))
	if _, err := svc2.ConfirmOrder(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if notified2 != 0 {
		t.Errorf("re-confirm must not notify, got %d", notified2)
	}
}

// notifierFunc adapts a function to the FulfillmentNotifier interface for tests.
type notifierFunc func(context.Context, *domain.Order) error

func (f notifierFunc) NotifyFulfilled(ctx context.Context, o *domain.Order) error { return f(ctx, o) }

func TestHandleWebhook_PropagatesSignatureError(t *testing.T) {
	repo := &fakeRepo{}
	reg := billing.NewRegistry(&fakeProvider{name: "paymob", err: billing.ErrInvalidSignature})
	svc := NewService(repo, reg)
	err := svc.HandleWebhook(context.Background(), uuid.New(), "paymob", http.Header{}, nil)
	if !errors.Is(err, billing.ErrInvalidSignature) {
		t.Errorf("want ErrInvalidSignature, got %v", err)
	}
	if repo.fulfillCalls != 0 {
		t.Error("must not fulfill on signature failure")
	}
}
