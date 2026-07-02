package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// MyFatoorahProvider integrates MyFatoorah (Kuwait/GCC). Checkout uses SendPayment to obtain a
// hosted InvoiceURL; the webhook is verified with HMAC-SHA256 (base64) over a documented,
// pipe-joined field ordering.
//
// VERIFY BEFORE PRODUCTION: base URL differs between the demo (apitest.myfatoorah.com) and live
// regional endpoints. Confirm the webhook secret and the exact field order below against your
// MyFatoorah portal — the signature is computed over the deposit/transaction fields they document.
type MyFatoorahProvider struct {
	apiToken      string
	webhookSecret string
	baseURL       string
	http          *http.Client
}

func NewMyFatoorahProvider(apiToken, webhookSecret, baseURL string) *MyFatoorahProvider {
	if baseURL == "" {
		baseURL = "https://api.myfatoorah.com"
	}
	return &MyFatoorahProvider{
		apiToken:      apiToken,
		webhookSecret: webhookSecret,
		baseURL:       baseURL,
		http:          &http.Client{Timeout: 20 * time.Second},
	}
}

func (m *MyFatoorahProvider) Name() string { return "myfatoorah" }

func (m *MyFatoorahProvider) CreateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error) {
	body := map[string]any{
		"InvoiceValue":       req.Amount,
		"DisplayCurrencyIso": req.Currency,
		"CustomerName":       req.TenantSlug,
		"CustomerEmail":      req.CustomerEmail,
		"CallBackUrl":        req.CallbackURL,
		"ErrorUrl":           req.CallbackURL,
		"CustomerReference":  req.TenantID,
		"NotificationOption": "LNK",
	}
	var out struct {
		IsSuccess bool `json:"IsSuccess"`
		Data      struct {
			InvoiceID  int64  `json:"InvoiceId"`
			InvoiceURL string `json:"InvoiceURL"`
		} `json:"Data"`
	}
	if err := m.postJSON(ctx, "/v2/SendPayment", body, &out); err != nil {
		return nil, err
	}
	if !out.IsSuccess || out.Data.InvoiceURL == "" {
		return nil, fmt.Errorf("myfatoorah: SendPayment did not return an invoice")
	}
	return &CheckoutResult{
		RedirectURL: out.Data.InvoiceURL,
		ProviderRef: strconv.FormatInt(out.Data.InvoiceID, 10),
	}, nil
}

// myfatoorahWebhook is the subset of MyFatoorah's webhook envelope we consume.
type myfatoorahWebhook struct {
	EventType int `json:"EventType"`
	Data      struct {
		InvoiceID          int64  `json:"InvoiceId"`
		TransactionStatus  string `json:"TransactionStatus"`
		PaymentID          string `json:"PaymentId"`
		InvoiceReference   string `json:"InvoiceReference"`
		InvoiceValue       string `json:"InvoiceValue"`
		Currency           string `json:"Currency"`
	} `json:"Data"`
}

func (m *MyFatoorahProvider) ParseAndVerifyWebhook(headers http.Header, body []byte) (*WebhookEvent, error) {
	sig := headers.Get("MyFatoorah-Signature")
	if sig == "" {
		sig = headers.Get("Myfatoorah-Signature")
	}
	if sig == "" {
		return nil, ErrInvalidSignature
	}

	var wh myfatoorahWebhook
	if err := json.Unmarshal(body, &wh); err != nil {
		return nil, err
	}

	// MyFatoorah signs a pipe-joined, alphabetically-keyed field string with HMAC-SHA256 (base64).
	d := wh.Data
	signable := fmt.Sprintf("Invoice.Id=%d,InvoiceReference=%s,InvoiceValue=%s,PaymentId=%s,TransactionStatus=%s",
		d.InvoiceID, d.InvoiceReference, d.InvoiceValue, d.PaymentID, d.TransactionStatus)

	mac := hmac.New(sha256.New, []byte(m.webhookSecret))
	mac.Write([]byte(signable))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return nil, ErrInvalidSignature
	}

	amount, _ := strconv.ParseFloat(d.InvoiceValue, 64)
	return &WebhookEvent{
		ProviderRef:    strconv.FormatInt(d.InvoiceID, 10),
		Paid:           d.TransactionStatus == "SUCCESS" || d.TransactionStatus == "Succss" || d.TransactionStatus == "Success",
		Amount:         amount,
		Currency:       d.Currency,
		IdempotencyKey: "myfatoorah:" + d.PaymentID,
		Raw:            body,
	}, nil
}

func (m *MyFatoorahProvider) postJSON(ctx context.Context, path string, body any, out any) error {
	buf, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiToken)
	resp, err := m.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("myfatoorah %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
