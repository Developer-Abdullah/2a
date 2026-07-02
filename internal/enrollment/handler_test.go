package enrollment

import (
	"bytes"
	"testing"

	"howett.net/plist"
)

func TestDeviceTypeFromProduct(t *testing.T) {
	cases := map[string]string{
		"iPhone13,2": "iphone",
		"iPhone15,3": "iphone",
		"iPad8,1":    "ipad",
		"iPad13,4":   "ipad",
		"Watch6,1":   "",
		"":           "",
	}
	for product, want := range cases {
		if got := deviceTypeFromProduct(product); got != want {
			t.Errorf("deviceTypeFromProduct(%q) = %q, want %q", product, got, want)
		}
	}
}

func TestSHA256HexNormalizesUDID(t *testing.T) {
	// Whitespace and case must not change the identity of a device.
	a := sha256Hex("00008030-000A1B2C3D")
	b := sha256Hex("  00008030-000a1b2c3d  ")
	if a != b {
		t.Errorf("expected normalized UDID hashes to match: %s vs %s", a, b)
	}
	if a == "" || len(a) != 64 {
		t.Errorf("unexpected hash form: %q", a)
	}
}

// TestDeviceAttributesUnmarshal proves the callback plist (the payload Apple posts inside its
// PKCS#7 envelope) decodes into the attributes we persist.
func TestDeviceAttributesUnmarshal(t *testing.T) {
	src := deviceAttributes{UDID: "00008030-ABC", Product: "iPhone14,5", Version: "17.5"}
	var buf bytes.Buffer
	if err := plist.NewEncoder(&buf).Encode(src); err != nil {
		t.Fatalf("encode: %v", err)
	}

	var got deviceAttributes
	if _, err := plist.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.UDID != src.UDID || got.Product != src.Product || got.Version != src.Version {
		t.Errorf("round-trip mismatch: got %+v want %+v", got, src)
	}
}
