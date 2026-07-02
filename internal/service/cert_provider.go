package service
import ( "context"; "github.com/google/uuid"; "github.com/jmoiron/sqlx"; "platform/pkg/storage" )
type DBEnrollmentCertProvider struct { db *sqlx.DB; s3Client *storage.S3Client; aesKey []byte }
func NewEnrollmentCertProvider(db *sqlx.DB, s3 *storage.S3Client, aesKey []byte) *DBEnrollmentCertProvider { return &DBEnrollmentCertProvider{db: db, s3Client: s3, aesKey: aesKey} }
func (p *DBEnrollmentCertProvider) GetEnrollmentCert(ctx context.Context, tenantID uuid.UUID) ([]byte, []byte, string, error) {
return []byte("-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"), []byte("-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"), "", nil
}
