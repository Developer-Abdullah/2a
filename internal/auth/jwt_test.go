package auth

import (
	"strings"
	"testing"
	"time"
)

const testPrivPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAKvqRKHPuQaFvkyEbBjn4IJ/Ko3SQMuQWaYGtOvsZGYoAoGCCqGSM49
AwEHoUQDQgAE+OjuUQUrCJrCQI43+k6jG4ABjnCA+3JMn2ka+ArK/1LAgrwFMymY
x3kLkjSr4yJRUth+6TarWwJmAXuht/mWEg==
-----END EC PRIVATE KEY-----`

const testPubPEM = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE+OjuUQUrCJrCQI43+k6jG4ABjnCA
+3JMn2ka+ArK/1LAgrwFMymYx3kLkjSr4yJRUth+6TarWwJmAXuht/mWEg==
-----END PUBLIC KEY-----`

func TestTokenRoundTrip(t *testing.T) {
	priv, err := ParsePrivateKey(testPrivPEM)
	if err != nil {
		t.Fatalf("ParsePrivateKey: %v", err)
	}
	pub, err := ParsePublicKey(testPubPEM)
	if err != nil {
		t.Fatalf("ParsePublicKey: %v", err)
	}

	tok, err := GenerateToken(CustomClaims{UserID: "u1", TenantID: "t1", Role: "user", DeviceID: "d1"}, priv, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims, err := ValidateToken(tok, pub)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "u1" || claims.DeviceID != "d1" || claims.Role != "user" {
		t.Errorf("unexpected claims: %+v", claims)
	}
}

// The .env / docker-compose env_file delivers keys with literal "\n"; normalizePEM must restore them.
func TestParseKeyWithEscapedNewlines(t *testing.T) {
	escaped := strings.ReplaceAll(testPrivPEM, "\n", `\n`)
	if _, err := ParsePrivateKey(escaped); err != nil {
		t.Errorf("private key with literal \\n should parse, got: %v", err)
	}
	escapedPub := strings.ReplaceAll(testPubPEM, "\n", `\n`)
	if _, err := ParsePublicKey(escapedPub); err != nil {
		t.Errorf("public key with literal \\n should parse, got: %v", err)
	}
}

func TestValidateTokenRejectsGarbage(t *testing.T) {
	pub, _ := ParsePublicKey(testPubPEM)
	if _, err := ValidateToken("not.a.jwt", pub); err == nil {
		t.Error("expected error for malformed token")
	}
}
