package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllowsUpToLimitThenBlocks(t *testing.T) {
	l := New(3, time.Minute)
	for i := 1; i <= 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Fatal("4th request should be blocked")
	}
	// A different key has its own budget.
	if !l.Allow("5.6.7.8") {
		t.Fatal("distinct key should be allowed")
	}
}

func TestLimiterResetsAfterWindow(t *testing.T) {
	l := New(1, time.Millisecond)
	if !l.Allow("k") {
		t.Fatal("first allowed")
	}
	if l.Allow("k") {
		t.Fatal("second blocked within window")
	}
	time.Sleep(2 * time.Millisecond)
	if !l.Allow("k") {
		t.Fatal("should be allowed after window resets")
	}
}
