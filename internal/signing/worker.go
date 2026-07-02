package signing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hibiken/asynq"

	"platform/pkg/crypto"
	"platform/pkg/storage"
	"platform/pkg/zsign"
)

type SigningContext struct {
	RawIPAS3Key, EncryptedP12S3Key, MobileProvisionS3Key, AppBundleID, AppVersion, AppName, TenantSlug, AppID, IconURL, FullSizeIconURL, CertID string
	EncryptedPassword                                                                                                                          []byte
}
type SigningRepository interface {
	GetSigningContext(ctx context.Context, jobID string, tenantID string) (*SigningContext, error)
	UpdateJobFailed(ctx context.Context, jobID, tenantID, errMsg string) error
	UpdateJobCompleted(ctx context.Context, jobID, tenantID, versionID, signedIPAKey, manifestKey string) error
}
type StorageClient interface {
	DownloadFile(ctx context.Context, key, destPath string) error
	UploadFile(ctx context.Context, key, filePath, contentType string) error
}

// Signer re-signs a raw IPA in place on the local filesystem. Implemented by the
// zsign adapter in production and mocked in tests.
type Signer interface {
	Sign(ctx context.Context, rawIPAPath, p12Path, provisionPath, password, outputPath string) error
}

// ZsignSigner adapts pkg/zsign to the Signer interface.
type ZsignSigner struct {
	Binary string // path to zsign binary; empty means look it up on PATH
}

func (z ZsignSigner) Sign(ctx context.Context, rawIPAPath, p12Path, provisionPath, password, outputPath string) error {
	s := &zsign.Signer{
		Binary:          z.Binary,
		RawIPA:          rawIPAPath,
		P12Cert:         p12Path,
		MobileProvision: provisionPath,
		Password:        password,
		OutputIPA:       outputPath,
	}
	return s.Sign(ctx)
}

type SigningProcessor struct {
	s3Client StorageClient
	repo     SigningRepository
	signer   Signer
	aesKey   []byte
	workDir  string // base directory for per-job scratch space; empty means the OS temp dir
}

func NewSigningProcessor(s3 StorageClient, repo SigningRepository, signer Signer, aesKey []byte, workDir string) *SigningProcessor {
	return &SigningProcessor{s3Client: s3, repo: repo, signer: signer, aesKey: aesKey, workDir: workDir}
}

// ProcessTask runs the real signing pipeline: it downloads the raw IPA plus the
// tenant's certificate assets, decrypts the .p12 and its password, invokes zsign,
// and uploads the signed artifact to a distinct S3 key before advancing the
// job/version state machine. Any failure marks the job failed with the reason so
// an Owner can inspect it — there is no silent success path.
func (p *SigningProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload IPASignPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	sctx, err := p.repo.GetSigningContext(ctx, payload.JobID, payload.TenantID)
	if err != nil {
		_ = p.repo.UpdateJobFailed(ctx, payload.JobID, payload.TenantID, "failed to load signing context: "+err.Error())
		return err
	}

	signedKey, err := p.sign(ctx, payload, sctx)
	if err != nil {
		_ = p.repo.UpdateJobFailed(ctx, payload.JobID, payload.TenantID, err.Error())
		return err
	}

	return p.repo.UpdateJobCompleted(ctx, payload.JobID, payload.TenantID, payload.VersionID, signedKey, "")
}

// sign performs the download/decrypt/zsign/upload sequence and returns the signed
// object's S3 key. All errors are returned so ProcessTask can record them on the job.
func (p *SigningProcessor) sign(ctx context.Context, payload IPASignPayload, sctx *SigningContext) (string, error) {
	if sctx.RawIPAS3Key == "" {
		return "", fmt.Errorf("version has no uploaded raw IPA to sign")
	}
	if sctx.EncryptedP12S3Key == "" || sctx.MobileProvisionS3Key == "" {
		return "", fmt.Errorf("tenant has no active signing certificate configured")
	}

	workspace, err := os.MkdirTemp(p.workDir, "sign-"+payload.JobID+"-")
	if err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	defer os.RemoveAll(workspace)

	rawPath := filepath.Join(workspace, "original.ipa")
	p12EncPath := filepath.Join(workspace, "cert.p12.enc")
	p12Path := filepath.Join(workspace, "cert.p12")
	provPath := filepath.Join(workspace, "app.mobileprovision")
	outPath := filepath.Join(workspace, "signed.ipa")

	if err := p.s3Client.DownloadFile(ctx, sctx.RawIPAS3Key, rawPath); err != nil {
		return "", fmt.Errorf("download raw IPA: %w", err)
	}
	if err := p.s3Client.DownloadFile(ctx, sctx.MobileProvisionS3Key, provPath); err != nil {
		return "", fmt.Errorf("download provisioning profile: %w", err)
	}

	// The .p12 is stored AES-256-GCM encrypted at rest; decrypt it into the workspace.
	if err := p.s3Client.DownloadFile(ctx, sctx.EncryptedP12S3Key, p12EncPath); err != nil {
		return "", fmt.Errorf("download certificate: %w", err)
	}
	p12Enc, err := os.ReadFile(p12EncPath)
	if err != nil {
		return "", fmt.Errorf("read encrypted certificate: %w", err)
	}
	p12Plain, err := crypto.Decrypt(p12Enc, p.aesKey)
	if err != nil {
		return "", fmt.Errorf("decrypt certificate: %w", err)
	}
	if err := os.WriteFile(p12Path, p12Plain, 0600); err != nil {
		return "", fmt.Errorf("write certificate: %w", err)
	}

	// The .p12 password is stored AES-256-GCM encrypted; an empty column means an
	// unprotected cert, which we allow.
	var password string
	if len(sctx.EncryptedPassword) > 0 {
		pw, err := crypto.Decrypt(sctx.EncryptedPassword, p.aesKey)
		if err != nil {
			return "", fmt.Errorf("decrypt certificate password: %w", err)
		}
		password = string(pw)
	}

	if err := p.signer.Sign(ctx, rawPath, p12Path, provPath, password, outPath); err != nil {
		return "", fmt.Errorf("zsign: %w", err)
	}

	builder := storage.NewS3PathBuilder(payload.TenantID)
	signedKey := builder.SignedIPA(sctx.AppID, payload.VersionID, payload.CertID)
	if err := p.s3Client.UploadFile(ctx, signedKey, outPath, "application/octet-stream"); err != nil {
		return "", fmt.Errorf("upload signed IPA: %w", err)
	}

	return signedKey, nil
}
