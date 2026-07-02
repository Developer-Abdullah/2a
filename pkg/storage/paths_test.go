package storage

import "testing"

func TestS3PathBuilder(t *testing.T) {
	b := NewS3PathBuilder("tenant1")
	cases := []struct{ got, want string }{
		{b.RawIPA("app1", "v1"), "tenant1/raw-ipa/app1/v1.ipa"},
		{b.SignedIPA("app1", "v1", "cert1"), "tenant1/signed-ipa/app1/v1-cert1.ipa"},
		{b.Manifest("app1", "v1"), "tenant1/manifests/app1/v1.plist"},
		{b.CertificateEncrypted("cert1"), "tenant1/certificates/cert1/cert.p12.enc"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("path = %q, want %q", c.got, c.want)
		}
	}
}
