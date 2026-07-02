package billing

import (
	"testing"
	"time"
)

func tp(t time.Time) *time.Time { return &t }

func TestSnapshotAllows_GracePeriod(t *testing.T) {
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	expired := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	cases := []struct {
		name string
		snap *SubscriptionSnapshot
		want bool
	}{
		{"nil snapshot blocked", nil, false},
		{"trial with no expiry allowed", &SubscriptionSnapshot{Status: "trialing"}, true},
		{"expired status with no expiry blocked", &SubscriptionSnapshot{Status: "expired"}, false},
		{"within paid period allowed", &SubscriptionSnapshot{Status: "active", ExpiresAt: tp(future)}, true},
		{"past expiry but inside grace allowed", &SubscriptionSnapshot{Status: "past_due", ExpiresAt: tp(expired), GraceUntil: tp(future)}, true},
		{"past expiry and past grace blocked", &SubscriptionSnapshot{Status: "past_due", ExpiresAt: tp(expired), GraceUntil: tp(expired)}, false},
		{"past expiry with no grace blocked", &SubscriptionSnapshot{Status: "active", ExpiresAt: tp(expired)}, false},
	}
	for _, tc := range cases {
		if got := snapshotAllows(tc.snap, now); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestRegistrySelectsProviders(t *testing.T) {
	reg := NewRegistry(
		NewPaymobProvider("k", "1", "2", "s", ""),
		NewMyFatoorahProvider("t", "s", ""),
	)
	if _, ok := reg.Get("paymob"); !ok {
		t.Error("paymob should be registered")
	}
	if _, ok := reg.Get("myfatoorah"); !ok {
		t.Error("myfatoorah should be registered")
	}
	if _, ok := reg.Get("stripe"); ok {
		t.Error("unregistered provider must not resolve")
	}
}
