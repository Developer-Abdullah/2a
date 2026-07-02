package billing

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
)

// buildSignedPaymobBody constructs a Paymob callback JSON whose hmac field is computed exactly the
// way the provider does, so we can prove verification accepts a well-formed event and rejects a
// tampered one — without calling Paymob.
func buildSignedPaymobBody(secret string, amountCents int64, success bool) []byte {
	created := "2026-07-02T12:00:00"
	currency := "EGP"
	txnID := int64(999001)
	orderID := int64(555)
	integrationID := int64(42)
	owner := int64(7)

	concat := fmt.Sprintf("%d", amountCents) + created + currency + "false" + "false" +
		fmt.Sprintf("%d", txnID) + fmt.Sprintf("%d", integrationID) + "false" + "false" + "false" + "false" +
		"false" + "false" + fmt.Sprintf("%d", orderID) + fmt.Sprintf("%d", owner) + "false" +
		"" + "" + "" + boolStr(success)

	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(concat))
	sig := hex.EncodeToString(mac.Sum(nil))

	body := fmt.Sprintf(`{"type":"TRANSACTION","obj":{
		"id":%d,"pending":false,"success":%t,"is_auth":false,"is_capture":false,
		"is_standalone_payment":false,"is_voided":false,"is_refunded":false,"is_3d_secure":false,
		"integration_id":%d,"has_parent_transaction":false,"error_occured":false,
		"amount_cents":%d,"currency":"%s","created_at":"%s","owner":%d,
		"order":{"id":%d},"source_data":{"pan":"","sub_type":"","type":""},"hmac":"%s"}}`,
		txnID, success, integrationID, amountCents, currency, created, owner, orderID, sig)
	return []byte(body)
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func TestPaymobWebhook_ValidSignature(t *testing.T) {
	p := NewPaymobProvider("k", "42", "1", "topsecret", "")
	body := buildSignedPaymobBody("topsecret", 500, true)

	ev, err := p.ParseAndVerifyWebhook(http.Header{}, body)
	if err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
	if !ev.Paid {
		t.Error("expected Paid=true for a successful transaction")
	}
	if ev.ProviderRef != "555" {
		t.Errorf("provider ref = %q, want 555", ev.ProviderRef)
	}
	if ev.Amount != 5.0 {
		t.Errorf("amount = %v, want 5.0", ev.Amount)
	}
}

func TestPaymobWebhook_TamperedSignatureRejected(t *testing.T) {
	p := NewPaymobProvider("k", "42", "1", "topsecret", "")
	// Signed with a different secret than the provider is configured with.
	body := buildSignedPaymobBody("attacker-secret", 500, true)

	if _, err := p.ParseAndVerifyWebhook(http.Header{}, body); err != ErrInvalidSignature {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}
