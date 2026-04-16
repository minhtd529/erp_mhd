package auth

import (
	"fmt"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
)

// TOTPKey holds the generated TOTP secret and the provisioning URL.
type TOTPKey struct {
	Secret string // base32-encoded secret; encrypt before storing
	URL    string // otpauth:// URL for QR code
}

// GenerateTOTPKey creates a new TOTP key for the given issuer and user account name.
// Uses SHA-1, 6 digits, 30-second period per RFC 6238.
func GenerateTOTPKey(issuer, email string) (*TOTPKey, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return nil, fmt.Errorf("totp: generate key: %w", err)
	}
	return &TOTPKey{
		Secret: key.Secret(),
		URL:    key.URL(),
	}, nil
}

// ValidateTOTP checks whether the given 6-digit code is valid for the base32 secret.
func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// QRCodePNG renders the otpauth URL as a PNG QR code and returns the raw bytes.
// size is the pixel width/height of the generated image (e.g., 256).
func QRCodePNG(url string, size int) ([]byte, error) {
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("totp: encode qr: %w", err)
	}
	return png, nil
}
