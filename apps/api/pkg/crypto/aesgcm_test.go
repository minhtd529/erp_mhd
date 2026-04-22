package crypto

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"testing"
)

// validTestKey returns a base64-encoded 32-byte test key.
func validTestKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestNewAESGCM_ValidKey(t *testing.T) {
	_, err := NewAESGCM(validTestKey(t))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNewAESGCM_EmptyKey(t *testing.T) {
	_, err := NewAESGCM("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
	if !errors.Is(err, ErrInvalidHRMKey) {
		t.Errorf("expected ErrInvalidHRMKey, got: %v", err)
	}
}

func TestNewAESGCM_ShortKey(t *testing.T) {
	short := base64.StdEncoding.EncodeToString([]byte("only16byteslong!"))
	_, err := NewAESGCM(short)
	if err == nil {
		t.Fatal("expected error for 16-byte key")
	}
	if !errors.Is(err, ErrInvalidHRMKey) {
		t.Errorf("expected ErrInvalidHRMKey, got: %v", err)
	}
}

func TestNewAESGCM_LongKey(t *testing.T) {
	long := base64.StdEncoding.EncodeToString(make([]byte, 64))
	_, err := NewAESGCM(long)
	if err == nil {
		t.Fatal("expected error for 64-byte key")
	}
	if !errors.Is(err, ErrInvalidHRMKey) {
		t.Errorf("expected ErrInvalidHRMKey, got: %v", err)
	}
}

func TestNewAESGCM_InvalidBase64(t *testing.T) {
	_, err := NewAESGCM("not!!valid!!base64@@")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
	if !errors.Is(err, ErrInvalidHRMKey) {
		t.Errorf("expected ErrInvalidHRMKey, got: %v", err)
	}
}

func TestAESGCM_EncryptDecryptRoundtrip(t *testing.T) {
	c, err := NewAESGCM(validTestKey(t))
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	cases := []string{
		"012345678901",
		"8123456789",
		"0123456789",
		"1234567890",
		"Vietcombank — Hà Nội — branch 1",
		"a",
	}
	for _, plain := range cases {
		enc, err := c.Encrypt(plain)
		if err != nil {
			t.Errorf("Encrypt(%q) error: %v", plain, err)
			continue
		}
		dec, err := c.Decrypt(enc)
		if err != nil {
			t.Errorf("Decrypt(Encrypt(%q)) error: %v", plain, err)
			continue
		}
		if dec != plain {
			t.Errorf("roundtrip mismatch: want %q, got %q", plain, dec)
		}
	}
}

func TestAESGCM_EncryptEmpty_ReturnsEmpty(t *testing.T) {
	c, _ := NewAESGCM(validTestKey(t))
	enc, err := c.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}
	if enc != "" {
		t.Errorf("expected empty string, got %q", enc)
	}
}

func TestAESGCM_DecryptEmpty_ReturnsEmpty(t *testing.T) {
	c, _ := NewAESGCM(validTestKey(t))
	dec, err := c.Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}
	if dec != "" {
		t.Errorf("expected empty string, got %q", dec)
	}
}

func TestAESGCM_DecryptTampered_ReturnsError(t *testing.T) {
	c, _ := NewAESGCM(validTestKey(t))
	enc, _ := c.Encrypt("sensitive data")

	// Flip last hex char
	tampered := enc[:len(enc)-1] + "f"
	if tampered == enc {
		tampered = enc[:len(enc)-1] + "0"
	}

	_, err := c.Decrypt(tampered)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
	if !errors.Is(err, ErrHRMDecryptFailed) {
		t.Errorf("expected ErrHRMDecryptFailed, got: %v", err)
	}
}

func TestAESGCM_DecryptTooShort_ReturnsError(t *testing.T) {
	c, _ := NewAESGCM(validTestKey(t))
	_, err := c.Decrypt("0102") // 2 bytes hex — way shorter than nonce
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
	if !errors.Is(err, ErrHRMCipherTooShort) {
		t.Errorf("expected ErrHRMCipherTooShort, got: %v", err)
	}
}

func TestAESGCM_TwoEncryptionsDiffer(t *testing.T) {
	c, _ := NewAESGCM(validTestKey(t))
	plain := "same plaintext"
	enc1, _ := c.Encrypt(plain)
	enc2, _ := c.Encrypt(plain)
	if enc1 == enc2 {
		t.Error("two encryptions of the same plaintext should produce different ciphertexts (random nonce)")
	}
	// Both should decrypt to the same plaintext
	dec1, _ := c.Decrypt(enc1)
	dec2, _ := c.Decrypt(enc2)
	if dec1 != plain || dec2 != plain {
		t.Errorf("both should decrypt to %q, got %q and %q", plain, dec1, dec2)
	}
}

func TestAESGCM_DecryptInvalidHex(t *testing.T) {
	c, _ := NewAESGCM(validTestKey(t))
	_, err := c.Decrypt("not-hex-at-all!!")
	if err == nil {
		t.Fatal("expected error for invalid hex input")
	}
	if !errors.Is(err, ErrHRMDecryptFailed) {
		t.Errorf("expected ErrHRMDecryptFailed, got: %v", err)
	}
}

func TestNewAESGCMFromEnv_MissingVar(t *testing.T) {
	old := os.Getenv("HRM_ENCRYPTION_KEY")
	os.Unsetenv("HRM_ENCRYPTION_KEY")
	defer func() {
		if old != "" {
			os.Setenv("HRM_ENCRYPTION_KEY", old)
		}
	}()

	_, err := NewAESGCMFromEnv()
	if err == nil {
		t.Fatal("expected error when HRM_ENCRYPTION_KEY not set")
	}
	if !strings.Contains(err.Error(), "HRM_ENCRYPTION_KEY") {
		t.Errorf("error should mention HRM_ENCRYPTION_KEY, got: %v", err)
	}
}

func TestNewAESGCMFromEnv_ValidVar(t *testing.T) {
	old := os.Getenv("HRM_ENCRYPTION_KEY")
	os.Setenv("HRM_ENCRYPTION_KEY", validTestKey(t))
	defer func() {
		if old != "" {
			os.Setenv("HRM_ENCRYPTION_KEY", old)
		} else {
			os.Unsetenv("HRM_ENCRYPTION_KEY")
		}
	}()

	c, err := NewAESGCMFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil cipher")
	}
}
