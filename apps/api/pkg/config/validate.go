package config

import (
	"fmt"
	"strings"
)

// ValidateProductionConfig ensures no insecure defaults leak into non-dev environments.
// Returns a single error listing ALL violations (not fail-fast).
// Call immediately after Load() and fatal on error.
func ValidateProductionConfig(cfg *Config) error {
	env := strings.ToLower(cfg.App.Env)
	if env == "development" || env == "dev" || env == "local" || env == "test" {
		return nil
	}

	var violations []string

	if cfg.JWT.Secret == "change-me-in-production" {
		violations = append(violations, "JWT_SECRET is set to default; override required")
	}

	if isAllZeroHex(cfg.TwoFA.EncryptionKey) {
		violations = append(violations, "TOTP_ENCRYPTION_KEY is all-zero hex; override with 32-byte random key")
	}

	if isAllZeroHex(cfg.HRM.BankEncryptionKey) {
		violations = append(violations, "HRM_BANK_ENCRYPTION_KEY is all-zero hex; override with 32-byte random key")
	}

	if cfg.MinIO.AccessKey == "minioadmin" {
		violations = append(violations, "MINIO_ACCESS_KEY is default 'minioadmin'; override required")
	}

	if cfg.MinIO.SecretKey == "minioadmin" {
		violations = append(violations, "MINIO_SECRET_KEY is default 'minioadmin'; override required")
	}

	if len(violations) > 0 {
		return fmt.Errorf(
			"insecure default config in env=%q (%d violations):\n  - %s",
			cfg.App.Env,
			len(violations),
			strings.Join(violations, "\n  - "),
		)
	}
	return nil
}

// isAllZeroHex reports whether s consists entirely of '0' characters.
// An empty string returns false — missing config is a separate concern.
func isAllZeroHex(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}
