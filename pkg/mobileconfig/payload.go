package mobileconfig

type OTAEnrollmentProfile struct {
PayloadContent      PayloadContent `plist:"PayloadContent"`
PayloadOrganization string         `plist:"PayloadOrganization"`
PayloadDisplayName  string         `plist:"PayloadDisplayName"`
PayloadVersion      int            `plist:"PayloadVersion"`
PayloadUUID         string         `plist:"PayloadUUID"`
PayloadIdentifier   string         `plist:"PayloadIdentifier"`
PayloadDescription  string         `plist:"PayloadDescription"`
PayloadType         string         `plist:"PayloadType"`
}

type PayloadContent struct {
URL              string   `plist:"URL"`
DeviceAttributes []string `plist:"DeviceAttributes"`
}
