package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

const upstreamCredentialEncryptionEnv = "UPSTREAM_CREDENTIAL_ENCRYPTION_KEY"

func UpstreamCredentialEncryptionConfigured() bool {
	return strings.TrimSpace(os.Getenv(upstreamCredentialEncryptionEnv)) != ""
}

func EncryptUpstreamCredential(plaintext []byte) (string, error) {
	aead, err := upstreamCredentialAEAD()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func DecryptUpstreamCredential(encoded string) ([]byte, error) {
	aead, err := upstreamCredentialAEAD()
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, errors.New("invalid encrypted upstream credential")
	}
	if len(ciphertext) < aead.NonceSize() {
		return nil, errors.New("invalid encrypted upstream credential")
	}
	nonce := ciphertext[:aead.NonceSize()]
	return aead.Open(nil, nonce, ciphertext[aead.NonceSize():], nil)
}

func upstreamCredentialAEAD() (cipher.AEAD, error) {
	secret := strings.TrimSpace(os.Getenv(upstreamCredentialEncryptionEnv))
	if secret == "" {
		return nil, errors.New(upstreamCredentialEncryptionEnv + " is not configured")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
