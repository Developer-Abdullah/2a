package signing

import (
	"context"
	"os"
	"testing"

	"github.com/hibiken/asynq"

	"platform/pkg/crypto"
)

// --- test doubles ---

type fakeStorage struct {
	objects     map[string][]byte // key -> bytes returned by DownloadFile
	uploaded    map[string]string // key -> local path passed to UploadFile
	uploadErr   error
	downloadErr error
}

func (f *fakeStorage) DownloadFile(_ context.Context, key, destPath string) error {
	if f.downloadErr != nil {
		return f.downloadErr
	}
	data, ok := f.objects[key]
	if !ok {
		return os.ErrNotExist
	}
	return os.WriteFile(destPath, data, 0600)
}

func (f *fakeStorage) UploadFile(_ context.Context, key, filePath, _ string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	if f.uploaded == nil {
		f.uploaded = map[string]string{}
	}
	f.uploaded[key] = filePath
	return nil
}

type fakeRepo struct {
	ctx         *SigningContext
	ctxErr      error
	failedMsg   string
	completedTo string // signed key recorded on completion
	completed   bool
}

func (r *fakeRepo) GetSigningContext(context.Context, string, string) (*SigningContext, error) {
	return r.ctx, r.ctxErr
}
func (r *fakeRepo) UpdateJobFailed(_ context.Context, _, _, msg string) error {
	r.failedMsg = msg
	return nil
}
func (r *fakeRepo) UpdateJobCompleted(_ context.Context, _, _, _, signedKey, _ string) error {
	r.completed = true
	r.completedTo = signedKey
	return nil
}

// fakeSigner writes a placeholder signed artifact so UploadFile has something to read.
type fakeSigner struct {
	called     bool
	gotPass    string
	returnErr  error
	gotP12Data []byte
}

func (s *fakeSigner) Sign(_ context.Context, _, p12Path, _, password, outputPath string) error {
	s.called = true
	s.gotPass = password
	s.gotP12Data, _ = os.ReadFile(p12Path)
	if s.returnErr != nil {
		return s.returnErr
	}
	return os.WriteFile(outputPath, []byte("SIGNED-IPA-BYTES"), 0600)
}

var testKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes

func mustTask(t *testing.T, p IPASignPayload) *asynq.Task {
	t.Helper()
	task, err := NewIPASignTask(p)
	if err != nil {
		t.Fatalf("NewIPASignTask: %v", err)
	}
	return task
}

func TestProcessTask_HappyPath(t *testing.T) {
	encP12, err := crypto.Encrypt([]byte("REAL-P12-CONTENT"), testKey)
	if err != nil {
		t.Fatalf("encrypt p12: %v", err)
	}
	encPass, err := crypto.Encrypt([]byte("s3cr3t"), testKey)
	if err != nil {
		t.Fatalf("encrypt password: %v", err)
	}

	storage := &fakeStorage{objects: map[string][]byte{
		"tenant/raw.ipa":       []byte("RAW-IPA"),
		"tenant/cert.p12.enc":  encP12,
		"tenant/app.mobileprov": []byte("PROVISION"),
	}}
	repo := &fakeRepo{ctx: &SigningContext{
		AppID:                "app-1",
		RawIPAS3Key:          "tenant/raw.ipa",
		EncryptedP12S3Key:    "tenant/cert.p12.enc",
		MobileProvisionS3Key: "tenant/app.mobileprov",
		EncryptedPassword:    encPass,
	}}
	signer := &fakeSigner{}
	p := NewSigningProcessor(storage, repo, signer, testKey, t.TempDir())

	err = p.ProcessTask(context.Background(), mustTask(t, IPASignPayload{JobID: "j1", VersionID: "v1", CertID: "c1", TenantID: "t1"}))
	if err != nil {
		t.Fatalf("ProcessTask returned error: %v (job failed msg: %q)", err, repo.failedMsg)
	}
	if !signer.called {
		t.Fatal("expected signer to be invoked")
	}
	if signer.gotPass != "s3cr3t" {
		t.Errorf("password not decrypted correctly: got %q", signer.gotPass)
	}
	if string(signer.gotP12Data) != "REAL-P12-CONTENT" {
		t.Errorf("p12 not decrypted correctly: got %q", signer.gotP12Data)
	}
	if !repo.completed {
		t.Fatal("expected job to be marked completed")
	}
	if _, ok := storage.uploaded[repo.completedTo]; !ok {
		t.Errorf("signed key %q was not uploaded", repo.completedTo)
	}
}

func TestProcessTask_NoCertificateFailsJob(t *testing.T) {
	repo := &fakeRepo{ctx: &SigningContext{
		AppID:       "app-1",
		RawIPAS3Key: "tenant/raw.ipa",
		// no cert keys -> must fail, not silently succeed
	}}
	signer := &fakeSigner{}
	p := NewSigningProcessor(&fakeStorage{objects: map[string][]byte{}}, repo, signer, testKey, t.TempDir())

	err := p.ProcessTask(context.Background(), mustTask(t, IPASignPayload{JobID: "j1", VersionID: "v1", TenantID: "t1"}))
	if err == nil {
		t.Fatal("expected error when no certificate is configured")
	}
	if signer.called {
		t.Error("signer must not run without a certificate")
	}
	if repo.completed {
		t.Error("job must not be completed when signing cannot proceed")
	}
	if repo.failedMsg == "" {
		t.Error("expected a failure reason to be recorded on the job")
	}
}

func TestProcessTask_SignerFailureMarksJobFailed(t *testing.T) {
	encP12, _ := crypto.Encrypt([]byte("P12"), testKey)
	storage := &fakeStorage{objects: map[string][]byte{
		"raw":  []byte("RAW"),
		"p12":  encP12,
		"prov": []byte("PROV"),
	}}
	repo := &fakeRepo{ctx: &SigningContext{
		AppID:                "app-1",
		RawIPAS3Key:          "raw",
		EncryptedP12S3Key:    "p12",
		MobileProvisionS3Key: "prov",
	}}
	signer := &fakeSigner{returnErr: os.ErrInvalid}
	p := NewSigningProcessor(storage, repo, signer, testKey, t.TempDir())

	err := p.ProcessTask(context.Background(), mustTask(t, IPASignPayload{JobID: "j1", VersionID: "v1", TenantID: "t1"}))
	if err == nil {
		t.Fatal("expected error when signer fails")
	}
	if repo.completed {
		t.Error("job must not be completed when signing fails")
	}
	if repo.failedMsg == "" {
		t.Error("expected failure reason recorded")
	}
}
