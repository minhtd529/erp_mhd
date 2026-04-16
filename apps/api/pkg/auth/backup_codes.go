package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	backupCodeChars  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no I, O, 0, 1 — ambiguous
	backupCodeLength = 8
	backupCodeCost   = 10 // faster than password cost=12; codes are high-entropy
)

// GenerateBackupCodes returns n random backup codes (uppercase alphanumeric, 8 chars each).
func GenerateBackupCodes(n int) ([]string, error) {
	codes := make([]string, n)
	for i := range codes {
		code, err := randomCode(backupCodeLength)
		if err != nil {
			return nil, fmt.Errorf("backup codes: generate: %w", err)
		}
		codes[i] = code
	}
	return codes, nil
}

// HashBackupCode returns a bcrypt hash of the given code (cost=10).
func HashBackupCode(code string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(strings.ToUpper(code)), backupCodeCost)
	if err != nil {
		return "", fmt.Errorf("backup codes: hash: %w", err)
	}
	return string(h), nil
}

// CheckBackupCode returns true if code matches the bcrypt hash.
func CheckBackupCode(code, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(strings.ToUpper(code))) == nil
}

// HashDeviceFingerprint returns the SHA-256 hex (64 chars) of the given fingerprint string.
// This is stored as device_id in the trusted_devices table.
func HashDeviceFingerprint(fingerprint string) string {
	h := sha256.Sum256([]byte(fingerprint))
	return hex.EncodeToString(h[:])
}

func randomCode(length int) (string, error) {
	alphabet := []rune(backupCodeChars)
	b := make([]rune, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
