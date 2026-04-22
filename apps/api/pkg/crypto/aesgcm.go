package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrInvalidHRMKey      = errors.New("crypto: invalid HRM encryption key")
	ErrHRMDecryptFailed   = errors.New("crypto: HRM decryption failed")
	ErrHRMCipherTooShort  = errors.New("crypto: HRM ciphertext too short")
)

// AESGCMCipher provides AES-256-GCM encrypt/decrypt for HRM PII fields.
// Key is sourced from HRM_ENCRYPTION_KEY env var (base64-encoded 32 bytes).
// Ciphertext is stored as hex-encoded string (compatible with TEXT columns).
type AESGCMCipher struct {
	aead cipher.AEAD
}

// NewAESGCMFromEnv reads HRM_ENCRYPTION_KEY from the environment and initialises
// the cipher. Returns an error if the variable is absent or invalid so callers
// can fail fast at startup.
func NewAESGCMFromEnv() (*AESGCMCipher, error) {
	return NewAESGCM(os.Getenv("HRM_ENCRYPTION_KEY"))
}

// NewAESGCM initialises a cipher from a base64-encoded 32-byte key string.
func NewAESGCM(base64Key string) (*AESGCMCipher, error) {
	if base64Key == "" {
		return nil, fmt.Errorf("%w: HRM_ENCRYPTION_KEY is not set", ErrInvalidHRMKey)
	}
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("%w: base64 decode failed: %v", ErrInvalidHRMKey, err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: must be 32 bytes, got %d", ErrInvalidHRMKey, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: aes.NewCipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: cipher.NewGCM: %w", err)
	}
	return &AESGCMCipher{aead: aead}, nil
}

// Encrypt encrypts plaintext with AES-256-GCM and returns a hex-encoded string
// containing [nonce || ciphertext || tag]. Empty plaintext returns "".
func (c *AESGCMCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: generate nonce: %w", err)
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(sealed), nil
}

// Decrypt decodes a hex-encoded AES-256-GCM ciphertext produced by Encrypt and
// returns the original plaintext. Empty input returns "".
func (c *AESGCMCipher) Decrypt(hexCiphertext string) (string, error) {
	if hexCiphertext == "" {
		return "", nil
	}
	data, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", fmt.Errorf("%w: hex decode: %v", ErrHRMDecryptFailed, err)
	}
	nonceSize := c.aead.NonceSize()
	if len(data) < nonceSize {
		return "", ErrHRMCipherTooShort
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrHRMDecryptFailed
	}
	return string(plaintext), nil
}
