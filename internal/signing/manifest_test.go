package signing

import (
	"strings"
	"testing"
)

func TestGenerateManifestPlist(t *testing.T) {
	out, err := GenerateManifestPlist("Cloud Notes", "com.demo.cloudnotes", "1.2.3",
		"https://api.example.com/v1/install/abc/download", "https://icon", "https://full")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	for _, want := range []string{
		"com.demo.cloudnotes",
		"Cloud Notes",
		"1.2.3",
		"https://api.example.com/v1/install/abc/download",
		"software-package",
		"<key>bundle-identifier</key>",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("manifest is missing %q", want)
		}
	}
	if !strings.HasPrefix(s, "<?xml") {
		t.Errorf("manifest should start with the XML declaration, got: %.20q", s)
	}
}
