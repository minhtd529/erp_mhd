package usecase_test

// ─── Mock implementations of domain repository interfaces ─────────────────────
// Kept package-private to test/ files only. No external mock framework needed.

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// ── mockUserRepo ──────────────────────────────────────────────────────────────

type mockUserRepo struct {
	user      *domain.UserForAuth
	findErr   error
	createID  uuid.UUID
	createErr error
}

func (m *mockUserRepo) FindByEmail(_ context.Context, _ string) (*domain.UserForAuth, error) {
	return m.user, m.findErr
}
func (m *mockUserRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.UserForAuth, error) {
	return m.user, m.findErr
}
func (m *mockUserRepo) CreateUser(_ context.Context, _ domain.CreateUserParams) (uuid.UUID, error) {
	return m.createID, m.createErr
}
func (m *mockUserRepo) UpdateLastLogin(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockUserRepo) UpdateUser(_ context.Context, _ domain.UpdateUserParams) error {
	return m.findErr
}
func (m *mockUserRepo) SoftDeleteUser(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return m.findErr
}
func (m *mockUserRepo) ListUsers(_ context.Context, _ domain.ListUsersFilter) ([]*domain.UserForAuth, int64, error) {
	if m.user != nil {
		return []*domain.UserForAuth{m.user}, 1, m.findErr
	}
	return []*domain.UserForAuth{}, 0, m.findErr
}

// ── mockRoleRepo ──────────────────────────────────────────────────────────────

type mockRoleRepo struct {
	role      *domain.Role
	findErr   error
	assignErr error
	roles     []string
	perms     []string
}

func (m *mockRoleRepo) FindByCode(_ context.Context, _ string) (*domain.Role, error) {
	return m.role, m.findErr
}
func (m *mockRoleRepo) AssignToUser(_ context.Context, _, _ uuid.UUID) error {
	return m.assignErr
}
func (m *mockRoleRepo) GetUserRoles(_ context.Context, _ uuid.UUID) ([]string, error) {
	return m.roles, nil
}
func (m *mockRoleRepo) GetUserPermissions(_ context.Context, _ uuid.UUID) ([]string, error) {
	return m.perms, nil
}

// ── mockTokenRepo ─────────────────────────────────────────────────────────────

type mockTokenRepo struct {
	stored    *domain.RefreshToken
	findErr   error
	createErr error
	revokeErr error
}

func (m *mockTokenRepo) CreateRefreshToken(_ context.Context, _ domain.RefreshToken) error {
	return m.createErr
}
func (m *mockTokenRepo) FindByHash(_ context.Context, _ string) (*domain.RefreshToken, error) {
	return m.stored, m.findErr
}
func (m *mockTokenRepo) Revoke(_ context.Context, _ string, _ time.Time) error {
	return m.revokeErr
}
func (m *mockTokenRepo) RevokeAllForUser(_ context.Context, _ uuid.UUID) error {
	return m.revokeErr
}

// ── mockJWTSvc ───────────────────────────────────────────────────────────────

type mockJWTSvc struct {
	accessToken string
	issueErr    error
	rawRefresh  string
	expiresAt   time.Time
	claims      *pkgauth.TokenClaims
	validateErr error
}

func (m *mockJWTSvc) IssueAccessToken(_ pkgauth.TokenClaims) (string, error) {
	return m.accessToken, m.issueErr
}
func (m *mockJWTSvc) IssueRefreshToken() (string, time.Time) {
	return m.rawRefresh, m.expiresAt
}
func (m *mockJWTSvc) ValidateAccessToken(_ string) (*pkgauth.TokenClaims, error) {
	return m.claims, m.validateErr
}
func (m *mockJWTSvc) AccessTokenTTLSeconds() int64 { return 900 }

// ── mockTwoFARepo ─────────────────────────────────────────────────────────────

type mockTwoFARepo struct {
	challenge       *domain.TwoFactorChallenge
	findChallengeErr error
	challengeCount  int  // incremented on IncrementChallengeAttempt
	encryptedSecret string
	backupCodes     []domain.BackupCode
	trustedDevice   *domain.TrustedDevice
	loginAttempts   int
}

func (m *mockTwoFARepo) SetTOTPSecret(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (m *mockTwoFARepo) GetTOTPSecret(_ context.Context, _ uuid.UUID) (string, error) {
	return m.encryptedSecret, nil
}
func (m *mockTwoFARepo) SetTwoFactorEnabled(_ context.Context, _ uuid.UUID, _ bool) error { return nil }
func (m *mockTwoFARepo) ClearTwoFactorSecret(_ context.Context, _ uuid.UUID) error        { return nil }

func (m *mockTwoFARepo) CreateChallenge(_ context.Context, _ domain.TwoFactorChallenge) error {
	return nil
}
func (m *mockTwoFARepo) FindChallenge(_ context.Context, _ string) (*domain.TwoFactorChallenge, error) {
	return m.challenge, m.findChallengeErr
}
func (m *mockTwoFARepo) IncrementChallengeAttempt(_ context.Context, _ string) (int, error) {
	m.challengeCount++
	return m.challengeCount, nil
}
func (m *mockTwoFARepo) InvalidateChallenge(_ context.Context, _ string) error { return nil }
func (m *mockTwoFARepo) MarkChallengeVerified(_ context.Context, _ string) error { return nil }

func (m *mockTwoFARepo) StoreBackupCodes(_ context.Context, _ uuid.UUID, _ []string) error { return nil }
func (m *mockTwoFARepo) GetUnusedBackupCodes(_ context.Context, _ uuid.UUID) ([]domain.BackupCode, error) {
	return m.backupCodes, nil
}
func (m *mockTwoFARepo) MarkBackupCodeUsed(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockTwoFARepo) DeleteAllBackupCodes(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockTwoFARepo) CountRemainingBackupCodes(_ context.Context, _ uuid.UUID) (int, error) {
	return len(m.backupCodes), nil
}

func (m *mockTwoFARepo) AddTrustedDevice(_ context.Context, _ domain.TrustedDevice) error { return nil }
func (m *mockTwoFARepo) FindTrustedDevice(_ context.Context, _ uuid.UUID, _ string) (*domain.TrustedDevice, error) {
	return m.trustedDevice, nil
}
func (m *mockTwoFARepo) CountTrustedDevices(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }
func (m *mockTwoFARepo) RevokeOldestTrustedDevice(_ context.Context, _ uuid.UUID) error  { return nil }
func (m *mockTwoFARepo) RevokeAllTrustedDevices(_ context.Context, _ uuid.UUID) error    { return nil }

func (m *mockTwoFARepo) IncrementLoginAttempts(_ context.Context, _ uuid.UUID) (int, error) {
	m.loginAttempts++
	return m.loginAttempts, nil
}
func (m *mockTwoFARepo) ResetLoginAttempts(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockTwoFARepo) LockAccount(_ context.Context, _ uuid.UUID, _ time.Time) error { return nil }
func (m *mockTwoFARepo) RespondToPushChallenge(_ context.Context, _ string, _ bool) error {
	return nil
}
func (m *mockTwoFARepo) FindPushChallenge(_ context.Context, _ string) (*domain.TwoFactorChallenge, error) {
	return m.challenge, m.findChallengeErr
}
