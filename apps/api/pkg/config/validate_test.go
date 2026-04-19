package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allDefaultConfig returns a Config pre-populated with every insecure default
// so tests can selectively override specific fields.
func allDefaultConfig(env string) *Config {
	return &Config{
		App: AppConfig{Env: env},
		JWT: JWTConfig{Secret: "change-me-in-production"},
		TwoFA: TwoFAConfig{
			EncryptionKey: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		HRM: HRMConfig{
			BankEncryptionKey: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		MinIO: MinIOConfig{
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
		},
	}
}

func TestValidateProductionConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cfg            *Config
		wantErr        bool
		wantViolations int       // expected count in error message
		wantSubstrings []string  // substrings that must appear in error
	}{
		{
			name:    "all defaults + ENV=development → no error (dev bypass)",
			cfg:     allDefaultConfig("development"),
			wantErr: false,
		},
		{
			name:           "all defaults + ENV=production → 5 violations",
			cfg:            allDefaultConfig("production"),
			wantErr:        true,
			wantViolations: 5,
			wantSubstrings: []string{
				"JWT_SECRET",
				"TOTP_ENCRYPTION_KEY",
				"HRM_BANK_ENCRYPTION_KEY",
				"MINIO_ACCESS_KEY",
				"MINIO_SECRET_KEY",
			},
		},
		{
			name: "JWT overridden + ENV=production → 4 violations (no JWT mention)",
			cfg: func() *Config {
				c := allDefaultConfig("production")
				c.JWT.Secret = "a-real-secret-that-is-long-enough-to-be-safe"
				return c
			}(),
			wantErr:        true,
			wantViolations: 4,
			wantSubstrings: []string{
				"TOTP_ENCRYPTION_KEY",
				"HRM_BANK_ENCRYPTION_KEY",
				"MINIO_ACCESS_KEY",
				"MINIO_SECRET_KEY",
			},
		},
		{
			name: "all secrets overridden + ENV=staging → no error",
			cfg: &Config{
				App: AppConfig{Env: "staging"},
				JWT: JWTConfig{Secret: "super-secret-jwt-key-for-staging-env"},
				TwoFA: TwoFAConfig{
					EncryptionKey: "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899",
				},
				HRM: HRMConfig{
					BankEncryptionKey: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				},
				MinIO: MinIOConfig{
					AccessKey: "prod-minio-access",
					SecretKey: "prod-minio-secret",
				},
			},
			wantErr: false,
		},
		{
			name: "TOTP key ends with non-zero digit → not all-zero → no violation for TOTP",
			cfg: func() *Config {
				c := allDefaultConfig("production")
				// Only TOTP is fixed; others remain default → 4 violations, TOTP not among them
				c.TwoFA.EncryptionKey = "00000000000000000000000000001234"
				return c
			}(),
			wantErr:        true,
			wantViolations: 4,
			wantSubstrings: []string{
				"JWT_SECRET",
				"HRM_BANK_ENCRYPTION_KEY",
				"MINIO_ACCESS_KEY",
				"MINIO_SECRET_KEY",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateProductionConfig(tc.cfg)

			if !tc.wantErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			msg := err.Error()

			// Verify violation count is mentioned in the error message
			if tc.wantViolations > 0 {
				assert.Contains(t, msg, fmt.Sprintf("%d violations", tc.wantViolations),
					"error should mention exact violation count")
			}

			// Verify each expected substring appears
			for _, sub := range tc.wantSubstrings {
				assert.True(t, strings.Contains(msg, sub),
					"error message should mention %q\ngot: %s", sub, msg)
			}
		})
	}
}

func TestIsAllZeroHex(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  bool
	}{
		{"0000000000000000000000000000000000000000000000000000000000000000", true},
		{"00000000000000000000000000001234", false},  // ends with non-zero
		{"", false},                                  // empty → not all-zero
		{"0", true},                                  // single zero
		{"a", false},                                 // non-zero non-numeric
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, isAllZeroHex(tc.input))
		})
	}
}
