package mobileconfig

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"howett.net/plist"
)

func BuildEnrollmentProfile(tenantName, callbackURL string) ([]byte, error) {
	payloadUUID := uuid.NewString()
	safeName := strings.ReplaceAll(strings.ToLower(tenantName), " ", "-")

	profile := OTAEnrollmentProfile{
		PayloadVersion: 1, PayloadUUID: payloadUUID, PayloadType: "Profile Service",
		PayloadIdentifier:   fmt.Sprintf("com.%s.enrollment.%s", safeName, payloadUUID),
		PayloadDisplayName:  fmt.Sprintf("%s Device Enrollment", tenantName),
		PayloadDescription:  "Secure authentication profile.",
		PayloadOrganization: tenantName,
		PayloadContent:      PayloadContent{URL: callbackURL, DeviceAttributes: []string{"UDID", "VERSION", "PRODUCT"}},
	}

	var buf bytes.Buffer
	encoder := plist.NewEncoder(&buf)
	encoder.Indent("    ")
	if err := encoder.Encode(profile); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GenerateAndSign(tenantName, callbackURL string, certPEM, keyPEM []byte, password string) ([]byte, error) {
	rawXML, err := BuildEnrollmentProfile(tenantName, callbackURL)
	if err != nil {
		return nil, err
	}
	return SignProfile(rawXML, certPEM, keyPEM, password)
}
