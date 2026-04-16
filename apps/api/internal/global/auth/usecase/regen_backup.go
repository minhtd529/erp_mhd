package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// RegenBackupCodesUseCase deletes old backup codes and issues a new set after
// verifying the user's password.
type RegenBackupCodesUseCase struct {
	users    domain.UserRepository
	twofa    domain.TwoFARepository
	auditLog *audit.Logger
}

func NewRegenBackupCodesUseCase(
	users domain.UserRepository,
	twofa domain.TwoFARepository,
	auditLog *audit.Logger,
) *RegenBackupCodesUseCase {
	return &RegenBackupCodesUseCase{users: users, twofa: twofa, auditLog: auditLog}
}

// Execute verifies password, deletes old codes, generates and stores new ones.
func (uc *RegenBackupCodesUseCase) Execute(ctx context.Context, userID uuid.UUID, password, ipAddress string) (*RegenBackupCodesResponse, error) {
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.TwoFactorEnabled {
		return nil, domain.ErrTwoFANotEnabled
	}

	if err := pkgauth.CheckPassword(password, user.HashedPassword); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Delete old codes
	if err := uc.twofa.DeleteAllBackupCodes(ctx, userID); err != nil {
		return nil, fmt.Errorf("regen backup: delete old: %w", err)
	}

	// Generate new codes
	rawCodes, err := pkgauth.GenerateBackupCodes(backupCodeCount)
	if err != nil {
		return nil, fmt.Errorf("regen backup: generate: %w", err)
	}

	hashes := make([]string, len(rawCodes))
	for i, c := range rawCodes {
		h, err := pkgauth.HashBackupCode(c)
		if err != nil {
			return nil, fmt.Errorf("regen backup: hash: %w", err)
		}
		hashes[i] = h
	}

	if err := uc.twofa.StoreBackupCodes(ctx, userID, hashes); err != nil {
		return nil, fmt.Errorf("regen backup: store: %w", err)
	}

	if uc.auditLog != nil {
		_ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     &userID,
			Module:     "global",
			Resource:   "users",
			ResourceID: &userID,
			Action:     "REGEN_BACKUP_CODES",
			IPAddress:  ipAddress,
		})
	}

	return &RegenBackupCodesResponse{
		BackupCodes:    rawCodes,
		RemainingCodes: len(rawCodes),
	}, nil
}
