package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	pkgcrypto "github.com/mdh/erp-audit/api/pkg/crypto"
)

// testEncKey is a 32-byte AES key for tests (not for production).
const testEncKey = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

// encryptTestSecret returns a ciphertext for the given TOTP secret using the testEncKey.
func encryptTestSecret(t *testing.T, secret string) string {
	t.Helper()
	c, err := pkgcrypto.Encrypt(testEncKey, secret)
	if err != nil {
		t.Fatalf("encrypt test secret: %v", err)
	}
	return c
}

// ─── Login with 2FA ──────────────────────────────────────────────────────────

func TestLoginUseCase_2FA(t *testing.T) {
	t.Parallel()

	hashed, _ := pkgauth.HashPassword("secret123")
	user2FA := &domain.UserForAuth{
		ID:               uuid.New(),
		Email:            "alice@example.com",
		HashedPassword:   hashed,
		Status:           "active",
		TwoFactorEnabled: true,
		TwoFactorMethod:  "totp",
		Roles:            []string{"AUDIT_STAFF"},
	}

	validJWT := &mockJWTSvc{
		accessToken: "access.token.here",
		rawRefresh:  uuid.New().String(),
		expiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	t.Run("2FA user without trusted device → 202 challenge returned", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{}
		uc := usecase.NewLoginUseCase(
			&mockUserRepo{user: user2FA},
			&mockRoleRepo{},
			&mockTokenRepo{},
			twoFARepo,
			validJWT,
			nil,
		)
		resp, err := uc.Execute(context.Background(), usecase.LoginRequest{
			Email:    "alice@example.com",
			Password: "secret123",
		}, "127.0.0.1")

		if !errors.Is(err, domain.ErrTwoFARequired) {
			t.Fatalf("want ErrTwoFARequired, got %v", err)
		}
		if resp == nil || resp.ChallengeID == "" {
			t.Error("expected non-empty challenge_id in response")
		}
		if resp.AccessToken != "" {
			t.Error("access_token must be empty when 2FA is required")
		}
	})

	t.Run("2FA user with valid trusted device → tokens issued directly", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{
			trustedDevice: &domain.TrustedDevice{
				ID:        uuid.New(),
				UserID:    user2FA.ID,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
		}
		uc := usecase.NewLoginUseCase(
			&mockUserRepo{user: user2FA},
			&mockRoleRepo{},
			&mockTokenRepo{},
			twoFARepo,
			validJWT,
			nil,
		)
		resp, err := uc.Execute(context.Background(), usecase.LoginRequest{
			Email:             "alice@example.com",
			Password:          "secret123",
			DeviceFingerprint: "some-device-fingerprint",
		}, "127.0.0.1")

		if err != nil {
			t.Fatalf("want nil error, got %v", err)
		}
		if resp.AccessToken == "" {
			t.Error("expected access_token when trusted device recognised")
		}
	})

	t.Run("wrong password → increments attempt count", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{}
		uc := usecase.NewLoginUseCase(
			&mockUserRepo{user: user2FA},
			&mockRoleRepo{},
			&mockTokenRepo{},
			twoFARepo,
			validJWT,
			nil,
		)
		_, err := uc.Execute(context.Background(), usecase.LoginRequest{
			Email:    "alice@example.com",
			Password: "wrongpassword",
		}, "127.0.0.1")

		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
		if twoFARepo.loginAttempts != 1 {
			t.Errorf("want 1 login attempt recorded, got %d", twoFARepo.loginAttempts)
		}
	})

	t.Run("account temporarily locked → ErrAccountLocked", func(t *testing.T) {
		t.Parallel()
		lockedUntil := time.Now().Add(10 * time.Minute)
		lockedUser := *user2FA
		lockedUser.LoginLockedUntil = &lockedUntil

		uc := usecase.NewLoginUseCase(
			&mockUserRepo{user: &lockedUser},
			&mockRoleRepo{},
			&mockTokenRepo{},
			&mockTwoFARepo{},
			validJWT,
			nil,
		)
		_, err := uc.Execute(context.Background(), usecase.LoginRequest{
			Email:    "alice@example.com",
			Password: "secret123",
		}, "127.0.0.1")

		if !errors.Is(err, domain.ErrAccountLocked) {
			t.Fatalf("want ErrAccountLocked, got %v", err)
		}
	})
}

// ─── Verify2FALogin ───────────────────────────────────────────────────────────

func TestVerify2FALoginUseCase(t *testing.T) {
	t.Parallel()

	// Generate a real TOTP key for testing
	key, err := pkgauth.GenerateTOTPKey("TestIssuer", "test@example.com")
	if err != nil {
		t.Fatalf("generate TOTP key: %v", err)
	}

	encSecret := encryptTestSecret(t, key.Secret)

	userID := uuid.New()
	user := &domain.UserForAuth{
		ID:               userID,
		Email:            "test@example.com",
		Roles:            []string{"AUDIT_STAFF"},
		TwoFactorEnabled: true,
	}

	validJWT := &mockJWTSvc{
		accessToken: "access.token.here",
		rawRefresh:  uuid.New().String(),
		expiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	activeChallenge := &domain.TwoFactorChallenge{
		ID:          uuid.New(),
		UserID:      userID,
		ChallengeID: "test-challenge-id",
		Method:      "totp",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	makeUC := func(twoFARepo *mockTwoFARepo) *usecase.Verify2FALoginUseCase {
		twoFARepo.encryptedSecret = encSecret
		return usecase.NewVerify2FALoginUseCase(
			&mockUserRepo{user: user},
			&mockTokenRepo{},
			twoFARepo,
			validJWT,
			testEncKey,
			5, 30, 5,
			nil,
		)
	}

	t.Run("expired challenge → ErrChallengeExpired", func(t *testing.T) {
		t.Parallel()
		expired := *activeChallenge
		expired.ExpiresAt = time.Now().Add(-1 * time.Minute)
		twoFARepo := &mockTwoFARepo{challenge: &expired}
		uc := makeUC(twoFARepo)

		_, err := uc.Execute(context.Background(), usecase.Verify2FALoginRequest{
			ChallengeID: "test-challenge-id",
			Code:        "000000",
		}, "127.0.0.1", "")

		if !errors.Is(err, domain.ErrChallengeExpired) {
			t.Fatalf("want ErrChallengeExpired, got %v", err)
		}
	})

	t.Run("invalidated challenge → ErrChallengeInvalidated", func(t *testing.T) {
		t.Parallel()
		invalidated := *activeChallenge
		now := time.Now()
		invalidated.InvalidatedAt = &now
		twoFARepo := &mockTwoFARepo{challenge: &invalidated}
		uc := makeUC(twoFARepo)

		_, err := uc.Execute(context.Background(), usecase.Verify2FALoginRequest{
			ChallengeID: "test-challenge-id",
			Code:        "000000",
		}, "127.0.0.1", "")

		if !errors.Is(err, domain.ErrChallengeInvalidated) {
			t.Fatalf("want ErrChallengeInvalidated, got %v", err)
		}
	})

	t.Run("wrong code → increments attempt count", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{challenge: activeChallenge}
		uc := makeUC(twoFARepo)

		_, err := uc.Execute(context.Background(), usecase.Verify2FALoginRequest{
			ChallengeID: "test-challenge-id",
			Code:        "000000", // deliberately wrong
		}, "127.0.0.1", "")

		if !errors.Is(err, domain.ErrTwoFAInvalid) {
			t.Fatalf("want ErrTwoFAInvalid, got %v", err)
		}
		if twoFARepo.challengeCount != 1 {
			t.Errorf("want 1 challenge attempt, got %d", twoFARepo.challengeCount)
		}
	})

	t.Run("5 wrong codes → ErrTooManyAttempts", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{
			challenge:      activeChallenge,
			challengeCount: 4, // already at 4; 5th attempt should trigger invalidation
		}
		uc := makeUC(twoFARepo)

		_, err := uc.Execute(context.Background(), usecase.Verify2FALoginRequest{
			ChallengeID: "test-challenge-id",
			Code:        "000000",
		}, "127.0.0.1", "")

		if !errors.Is(err, domain.ErrTooManyAttempts) {
			t.Fatalf("want ErrTooManyAttempts, got %v", err)
		}
	})

	t.Run("challenge not found → ErrChallengeNotFound", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{
			findChallengeErr: domain.ErrChallengeNotFound,
		}
		uc := makeUC(twoFARepo)

		_, err := uc.Execute(context.Background(), usecase.Verify2FALoginRequest{
			ChallengeID: "nonexistent",
			Code:        "000000",
		}, "127.0.0.1", "")

		if !errors.Is(err, domain.ErrChallengeNotFound) {
			t.Fatalf("want ErrChallengeNotFound, got %v", err)
		}
	})
}

// ─── VerifyBackupCode ─────────────────────────────────────────────────────────

func TestVerifyBackupCodeUseCase(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	user := &domain.UserForAuth{
		ID:               userID,
		Email:            "test@example.com",
		TwoFactorEnabled: true,
	}

	// Generate a real backup code and its hash
	rawCode := "ABCD1234"
	hash, err := pkgauth.HashBackupCode(rawCode)
	if err != nil {
		t.Fatalf("hash backup code: %v", err)
	}

	activeChallenge := &domain.TwoFactorChallenge{
		ID:          uuid.New(),
		UserID:      userID,
		ChallengeID: "backup-challenge-id",
		Method:      "totp",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	validJWT := &mockJWTSvc{
		accessToken: "access.token.here",
		rawRefresh:  uuid.New().String(),
		expiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	t.Run("valid backup code → tokens issued", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{
			challenge: activeChallenge,
			backupCodes: []domain.BackupCode{
				{ID: uuid.New(), UserID: userID, CodeHash: hash},
			},
		}
		uc := usecase.NewVerifyBackupCodeUseCase(
			&mockUserRepo{user: user},
			&mockTokenRepo{},
			twoFARepo,
			validJWT,
			nil,
		)

		resp, err := uc.Execute(context.Background(), usecase.VerifyBackupCodeRequest{
			ChallengeID: "backup-challenge-id",
			Code:        rawCode,
		}, "127.0.0.1")

		if err != nil {
			t.Fatalf("want nil error, got %v", err)
		}
		if resp.AccessToken == "" {
			t.Error("expected non-empty access_token")
		}
	})

	t.Run("invalid backup code → ErrBackupCodeInvalid", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{
			challenge: activeChallenge,
			backupCodes: []domain.BackupCode{
				{ID: uuid.New(), UserID: userID, CodeHash: hash},
			},
		}
		uc := usecase.NewVerifyBackupCodeUseCase(
			&mockUserRepo{user: user},
			&mockTokenRepo{},
			twoFARepo,
			validJWT,
			nil,
		)

		_, err := uc.Execute(context.Background(), usecase.VerifyBackupCodeRequest{
			ChallengeID: "backup-challenge-id",
			Code:        "WRONGCOD",
		}, "127.0.0.1")

		if !errors.Is(err, domain.ErrBackupCodeInvalid) {
			t.Fatalf("want ErrBackupCodeInvalid, got %v", err)
		}
	})

	t.Run("no remaining backup codes → ErrBackupCodeInvalid", func(t *testing.T) {
		t.Parallel()
		twoFARepo := &mockTwoFARepo{
			challenge:   activeChallenge,
			backupCodes: []domain.BackupCode{}, // empty
		}
		uc := usecase.NewVerifyBackupCodeUseCase(
			&mockUserRepo{user: user},
			&mockTokenRepo{},
			twoFARepo,
			validJWT,
			nil,
		)

		_, err := uc.Execute(context.Background(), usecase.VerifyBackupCodeRequest{
			ChallengeID: "backup-challenge-id",
			Code:        rawCode,
		}, "127.0.0.1")

		if !errors.Is(err, domain.ErrBackupCodeInvalid) {
			t.Fatalf("want ErrBackupCodeInvalid, got %v", err)
		}
	})
}
