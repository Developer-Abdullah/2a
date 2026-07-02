package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// PaymobProvider integrates Paymob's Accept API (Egypt). Checkout is Paymob's 3-step handshake
// (auth token -> order -> payment key -> hosted iframe). The webhook is verified with HMAC-SHA512
// over a fixed, provider-documented field ordering.
//
// VERIFY BEFORE PRODUCTION: endpoints, integration/iframe ids, and the HMAC field list are taken
// from Paymob's documented Accept flow. Confirm them against your live dashboard — Paymob has
// region-specific base URLs and occasionally revises the callback field set.
type PaymobProvider struct {
	apiKey        string
	integrationID string
	iframeID      string
	hmacSecret    string
	baseURL       string
	http          *http.Client
}

func NewPaymobProvider(apiKey, integrationID, iframeID, hmacSecret, baseURL string) *PaymobProvider {
	if baseURL == "" {
		baseURL = "https://accept.paymob.com"
	}
	return &PaymobProvider{
		apiKey:        apiKey,
		integrationID: integrationID,
		iframeID:      iframeID,
		hmacSecret:    hmacSecret,
		baseURL:       baseURL,
		http:          &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *PaymobProvider) Name() string { return "paymob" }

func (p *PaymobProvider) CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error) {
	amountCents := int64(req.Amount*100 + 0.5)

	authToken, err := p.authToken(ctx)
	if err != nil {
		return nil, err
	}

	orderID, err := p.createOrder(ctx, authToken, amountCents, req.Currency)
	if err != nil {
		return nil, err
	}

	payToken, err := p.paymentKey(ctx, authToken, orderID, amountCents, req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/acceptance/iframes/%s?payment_token=%s", p.baseURL, p.iframeID, payToken)
	return &CheckoutResult{RedirectURL: url, ProviderRef: strconv.FormatInt(orderID, 10)}, nil
}

func (p *PaymobProvider) authToken(ctx context.Context) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	if err := p.postJSON(ctx, "/api/auth/tokens", map[string]any{"api_key": p.apiKey}, &out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("paymob: empty auth token")
	}
	return out.Token, nil
}

func (p *PaymobProvider) createOrder(ctx context.Context, authToken string, amountCents int64, currency string) (int64, error) {
	var out struct {
		ID int64 `json:"id"`
	}
	body := map[string]any{
		"auth_token":      authToken,
		"delivery_needed": false,
		"amount_cents":    amountCents,
		"currency":        currency,
		"items":           []any{},
	}
	if err := p.postJSON(ctx, "/api/ecommerce/orders", body, &out); err != nil {
		return 0, err
	}
	return out.ID, nil
}

func (p *PaymobProvider) paymentKey(ctx context.Context, authToken string, orderID, amountCents int64, req CheckoutRequest) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	body := map[string]any{
		"auth_token":   authToken,
		"amount_cents": amountCents,
		"expiration":   3600,
		"order_id":     orderID,
		"currency":     req.Currency,
		"integration_id": p.integrationID,
		"billing_data": map[string]any{
			"email":        firstNonEmpty(req.CustomerEmail, "billing@"+req.TenantSlug+".local"),
			"first_name":   req.TenantSlug,
			"last_name":    "merchant",
			"phone_number": "NA",
			"country":      "NA", "city": "NA", "street": "NA", "building": "NA",
			"floor": "NA", "apartment": "NA",
		},
	}
	if err := p.postJSON(ctx, "/api/acceptance/payment_keys", body, &out); err != nil {
		return "", err
	}
	return out.Token, nil
}

// paymobCallback is the subset of Paymob's transaction webhook we consume.
type paymobCallback struct {
	Type string `json:"type"`
	Obj  struct {
		ID       int64  `json:"id"`
		Pending  bool   `json:"pending"`
		Success  bool   `json:"success"`
		IsAuth   bool   `json:"is_auth"`
		IsCapture bool  `json:"is_capture"`
		IsStandalonePayment bool `json:"is_standalone_payment"`
		IsVoided bool   `json:"is_voided"`
		IsRefunded bool `json:"is_refunded"`
		Is3DSecure bool `json:"is_3d_secure"`
		IntegrationID int64 `json:"integration_id"`
		HasParentTransaction bool `json:"has_parent_transaction"`
		ErrorOccured bool `json:"error_occured"`
		AmountCents int64  `json:"amount_cents"`
		Currency    string `json:"currency"`
		CreatedAt   string `json:"created_at"`
		Owner       int64  `json:"owner"`
		Order       struct {
			ID int64 `json:"id"`
		} `json:"order"`
		SourceData struct {
			Pan     string `json:"pan"`
			SubType string `json:"sub_type"`
			Type    string `json:"type"`
		} `json:"source_data"`
		HMAC string `json:"hmac"`
	} `json:"obj"`
}

func (p *PaymobProvider) ParseAndVerifyWebhook(headers http.Header, body []byte) (*WebhookEvent, error) {
	var cb paymobCallback
	if err := json.Unmarshal(body, &cb); err != nil {
		return nil, err
	}

	// Paymob concatenates these fields, in this exact order, then HMAC-SHA512s them with the secret.
	o := cb.Obj
	concat := b(o.AmountCents) + o.CreatedAt + o.Currency + tf(o.ErrorOccured) + tf(o.HasParentTransaction) +
		b(o.ID) + b(o.IntegrationID) + tf(o.Is3DSecure) + tf(o.IsAuth) + tf(o.IsCapture) + tf(o.IsRefunded) +
		tf(o.IsStandalonePayment) + tf(o.IsVoided) + b(o.Order.ID) + b(o.Owner) + tf(o.Pending) +
		o.SourceData.Pan + o.SourceData.SubType + o.SourceData.Type + tf(o.Success)

	mac := hmac.New(sha512.New, []byte(p.hmacSecret))
	mac.Write([]byte(concat))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(o.HMAC)) {
		return nil, ErrInvalidSignature
	}

	return &WebhookEvent{
		ProviderRef:    strconv.FormatInt(o.Order.ID, 10),
		Paid:           o.Success && !o.IsRefunded && !o.IsVoided && !o.ErrorOccured,
		Amount:         float64(o.AmountCents) / 100.0,
		Currency:       o.Currency,
		IdempotencyKey: "paymob:" + strconv.FormatInt(o.ID, 10),
		Raw:            body,
	}, nil
}

func (p *PaymobProvider) postJSON(ctx context.Context, path string, body any, out any) error {
	buf, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("paymob %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// b renders an integer the way Paymob does in its HMAC string.
func b(n int64) string { return strconv.FormatInt(n, 10) }

// tf renders a bool as the lowercase "true"/"false" Paymob uses in its HMAC string.
func tf(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
