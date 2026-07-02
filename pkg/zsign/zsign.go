package zsign

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Signer holds the fully-resolved local filesystem paths that a single zsign
// invocation needs. All paths must already exist on disk; the caller owns their
// lifecycle (creation and cleanup).
type Signer struct {
	Binary          string // path to the zsign binary; defaults to "zsign" when empty
	RawIPA          string // input .ipa to be re-signed
	P12Cert         string // decrypted .p12 (key + cert bundle)
	MobileProvision string // .mobileprovision profile
	Password        string // cleartext password for the .p12
	OutputIPA       string // destination path for the signed .ipa
}

// Sign runs zsign against the configured inputs, producing OutputIPA.
//
// The command is built as an explicit argv slice and executed WITHOUT a shell,
// so tenant-controlled filenames cannot inject shell metacharacters. The p12
// password is passed via zsign's -p flag rather than embedded in a path.
func (s *Signer) Sign(ctx context.Context) error {
	bin := s.Binary
	if bin == "" {
		bin = "zsign"
	}

	// zsign CLI: zsign -k <p12> -p <password> -m <prov> -o <output> <input.ipa>
	args := []string{"-k", s.P12Cert, "-m", s.MobileProvision, "-o", s.OutputIPA}
	if s.Password != "" {
		args = append(args, "-p", s.Password)
	}
	args = append(args, s.RawIPA)

	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zsign failed: %w | stderr: %s", err, stderr.String())
	}
	return nil
}
