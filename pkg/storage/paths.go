package storage
import "fmt"

type S3PathBuilder struct { TenantID string }
func NewS3PathBuilder(tenantID string) *S3PathBuilder { return &S3PathBuilder{TenantID: tenantID} }
func (p *S3PathBuilder) RawIPA(appID, versionID string) string { return fmt.Sprintf("%s/raw-ipa/%s/%s.ipa", p.TenantID, appID, versionID) }
func (p *S3PathBuilder) SignedIPA(appID, versionID, certID string) string { return fmt.Sprintf("%s/signed-ipa/%s/%s-%s.ipa", p.TenantID, appID, versionID, certID) }
func (p *S3PathBuilder) Manifest(appID, versionID string) string { return fmt.Sprintf("%s/manifests/%s/%s.plist", p.TenantID, appID, versionID) }
func (p *S3PathBuilder) CertificateEncrypted(certID string) string { return fmt.Sprintf("%s/certificates/%s/cert.p12.enc", p.TenantID, certID) }
func (p *S3PathBuilder) ExportCSV(adminID, timestamp string) string { return fmt.Sprintf("%s/exports/%s/%s-export.csv", p.TenantID, adminID, timestamp) }
