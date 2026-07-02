package crypto

import (
"crypto/aes"
"crypto/cipher"
"crypto/rand"
"errors"
"io"
)

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
if len(key) != 32 { return nil, errors.New("key must be 32 bytes") }
block, err := aes.NewCipher(key)
if err != nil { return nil, err }
aesGCM, err := cipher.NewGCM(block)
if err != nil { return nil, err }
nonce := make([]byte, aesGCM.NonceSize())
if _, err := io.ReadFull(rand.Reader, nonce); err != nil { return nil, err }
return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
if len(key) != 32 { return nil, errors.New("key must be 32 bytes") }
block, err := aes.NewCipher(key)
if err != nil { return nil, err }
aesGCM, err := cipher.NewGCM(block)
if err != nil { return nil, err }
nonceSize := aesGCM.NonceSize()
if len(ciphertext) < nonceSize { return nil, errors.New("ciphertext too short") }
nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
return aesGCM.Open(nil, nonce, encryptedData, nil)
}
