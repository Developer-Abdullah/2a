package signing

import (
	"bytes"
	"fmt"
	"text/template"
)

// appleManifestTemplate dictates the exact XML structure mandated by iOS for OTA installations via itms-services://
// DO NOT alter the XML nodes, as iOS validation is incredibly strict and will silently fail.
const appleManifestTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>items</key>
	<array>
		<dict>
			<key>assets</key>
			<array>
				<dict>
					<key>kind</key>
					<string>software-package</string>
					<key>url</key>
					<string>{{.APIDownloadURL}}</string>
				</dict>
				<dict>
					<key>kind</key>
					<string>display-image</string>
					<key>needs-shine</key>
					<false/>
					<key>url</key>
					<string>{{.IconURL}}</string>
				</dict>
				<dict>
					<key>kind</key>
					<string>full-size-image</string>
					<key>needs-shine</key>
					<false/>
					<key>url</key>
					<string>{{.FullSizeIconURL}}</string>
				</dict>
			</array>
			<key>metadata</key>
			<dict>
				<key>bundle-identifier</key>
				<string>{{.BundleID}}</string>
				<key>bundle-version</key>
				<string>{{.Version}}</string>
				<key>kind</key>
				<string>software</string>
				<key>title</key>
				<string>{{.AppName}}</string>
			</dict>
		</dict>
	</array>
</dict>
</plist>`

// manifestData holds the parameters injected into the Apple XML template.
type manifestData struct {
	AppName         string
	BundleID        string
	Version         string
	APIDownloadURL  string
	IconURL         string
	FullSizeIconURL string
}

// GenerateManifestPlist dynamically constructs the Apple required property list for OTA app distribution.
// CRITICAL: The apiDownloadURL MUST point to the Core API's secure redirect endpoint, not the S3 bucket directly.
func GenerateManifestPlist(appName, bundleID, version, apiDownloadURL, iconURL, fullSizeIconURL string) ([]byte, error) {
	data := manifestData{
		AppName:         appName,
		BundleID:        bundleID,
		Version:         version,
		APIDownloadURL:  apiDownloadURL,
		IconURL:         iconURL,
		FullSizeIconURL: fullSizeIconURL,
	}

	// We use text/template because HTML template escaping rules can corrupt the strict DTD format if not managed perfectly.
	tmpl, err := template.New("manifest").Parse(appleManifestTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Apple manifest template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to generate manifest XML: %w", err)
	}

	return buf.Bytes(), nil
}