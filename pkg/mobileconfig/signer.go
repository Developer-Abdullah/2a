package mobileconfig

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"go.mozilla.org/pkcs7"
)

func SignProfile(profileXML []byte, certPEM []byte, keyPEM []byte, password string) ([]byte, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, errors.New("failed to decode cert")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, errors.New("failed to decode private key")
	}

	if password != "" || keyBlock.Type == "ENCRYPTED PRIVATE KEY" || isLegacyEncryptedPEMBlock(keyBlock) {
		return nil, errors.New("encrypted private keys are not supported; provide an unencrypted PKCS#8, RSA, or EC private key")
	}

	var privateKey crypto.PrivateKey
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	default:
		privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	}
	if err != nil {
		return nil, err
	}

	signedData, err := pkcs7.NewSignedData(profileXML)
	if err != nil {
		return nil, err
	}
	if err := signedData.AddSigner(cert, privateKey, pkcs7.SignerInfoConfig{}); err != nil {
		return nil, err
	}
	return signedData.Finish()
}

func isLegacyEncryptedPEMBlock(block *pem.Block) bool {
	_, hasDEKInfo := block.Headers["DEK-Info"]
	procType, hasProcType := block.Headers["Proc-Type"]
	return hasDEKInfo || (hasProcType && procType == "4,ENCRYPTED")
}
