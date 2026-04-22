package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/crypto"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// SensitiveFieldsResponse is returned by GET /hrm/employees/:id/sensitive.
// All encrypted fields are decrypted plaintext values.
type SensitiveFieldsResponse struct {
	ID              string  `json:"id"`
	EmployeeCode    string  `json:"employee_code,omitempty"`
	FullName        string  `json:"full_name"`
	CCCD            *string `json:"cccd,omitempty"`
	CCCDIssuedDate  *string `json:"cccd_issued_date,omitempty"`
	CCCDIssuedPlace *string `json:"cccd_issued_place,omitempty"`
	PassportNumber  *string `json:"passport_number,omitempty"`
	PassportExpiry  *string `json:"passport_expiry,omitempty"`
	MSTCaNhan       *string `json:"mst_ca_nhan,omitempty"`
	SoBHXH          *string `json:"so_bhxh,omitempty"`
	BankAccount     *string `json:"bank_account,omitempty"`
	BankName        *string `json:"bank_name,omitempty"`
	BankBranch      *string `json:"bank_branch,omitempty"`
	AccessedAt      string  `json:"accessed_at"`
}

// UpdateSensitiveRequest is the body for PUT /hrm/employees/:id/sensitive.
type UpdateSensitiveRequest struct {
	CCCD            *string `json:"cccd,omitempty"`
	CCCDIssuedDate  *string `json:"cccd_issued_date,omitempty"`
	CCCDIssuedPlace *string `json:"cccd_issued_place,omitempty"`
	PassportNumber  *string `json:"passport_number,omitempty"`
	PassportExpiry  *string `json:"passport_expiry,omitempty"`
	MSTCaNhan       *string `json:"mst_ca_nhan,omitempty"`
	SoBHXH          *string `json:"so_bhxh,omitempty"`
	BankAccount     *string `json:"bank_account,omitempty"`
	BankName        *string `json:"bank_name,omitempty"`
	BankBranch      *string `json:"bank_branch,omitempty"`
}

// ─── UseCase ─────────────────────────────────────────────────────────────────

type SensitiveUseCase struct {
	repo     domain.EmployeeRepository
	cipher   *crypto.AESGCMCipher
	auditLog *audit.Logger
}

func NewSensitiveUseCase(repo domain.EmployeeRepository, cipher *crypto.AESGCMCipher, auditLog *audit.Logger) *SensitiveUseCase {
	return &SensitiveUseCase{repo: repo, cipher: cipher, auditLog: auditLog}
}

// hasSensitivePermission returns true only for roles allowed to decrypt PII.
// Per SPEC §15 and permission matrix: SUPER_ADMIN, CHAIRMAN, CEO, HR_MANAGER.
func hasSensitivePermission(roles []string) bool {
	for _, r := range roles {
		switch r {
		case "SUPER_ADMIN", "CHAIRMAN", "CEO", "HR_MANAGER":
			return true
		}
	}
	return false
}

// GetSensitive decrypts and returns PII fields. Writes EMPLOYEE_PII_ACCESSED
// audit log before returning data — fails closed if audit write fails.
func (uc *SensitiveUseCase) GetSensitive(
	ctx context.Context,
	employeeID uuid.UUID,
	callerID uuid.UUID,
	callerRoles []string,
	ip string,
) (*SensitiveFieldsResponse, error) {
	if !hasSensitivePermission(callerRoles) {
		return nil, domain.ErrInsufficientPermission
	}

	e, err := uc.repo.FindByID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("sensitive.GetSensitive: %w", err)
	}

	decryptField := func(stored *string, fieldName string) *string {
		if stored == nil || *stored == "" {
			return nil
		}
		plain, err := uc.cipher.Decrypt(*stored)
		if err != nil {
			log.Printf("ERROR sensitive.GetSensitive decrypt %s employee=%s: %v", fieldName, employeeID, err)
			return nil
		}
		return &plain
	}

	now := time.Now().UTC()

	// Write audit log BEFORE returning data — fail closed on audit failure.
	_, auditErr := uc.auditLog.Log(ctx, audit.Entry{
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "employees",
		ResourceID: &employeeID,
		Action:     "EMPLOYEE_PII_ACCESSED",
		NewValue:   map[string]any{"fields_read": []string{"cccd", "mst_ca_nhan", "so_bhxh", "bank_account"}},
		IPAddress:  ip,
	})
	if auditErr != nil {
		log.Printf("ERROR sensitive.GetSensitive audit EMPLOYEE_PII_ACCESSED employee=%s caller=%s: %v",
			employeeID, callerID, auditErr)
		return nil, fmt.Errorf("audit log write failed — access denied")
	}

	empCode := ""
	if e.EmployeeCode != nil {
		empCode = *e.EmployeeCode
	}

	return &SensitiveFieldsResponse{
		ID:              e.ID.String(),
		EmployeeCode:    empCode,
		FullName:        e.FullName,
		CCCD:            decryptField(e.CccdEncrypted, "cccd"),
		CCCDIssuedDate:  dateStr(e.CccdIssuedDate),
		CCCDIssuedPlace: e.CccdIssuedPlace,
		PassportNumber:  e.PassportNumber,
		PassportExpiry:  dateStr(e.PassportExpiry),
		MSTCaNhan:       decryptField(e.MstCaNhanEncrypted, "mst_ca_nhan"),
		SoBHXH:          decryptField(e.SoBhxhEncrypted, "so_bhxh"),
		BankAccount:     decryptField(e.BankAccountEncrypted, "bank_account"),
		BankName:        e.BankName,
		BankBranch:      e.BankBranch,
		AccessedAt:      now.Format(time.RFC3339),
	}, nil
}

// UpdateSensitive encrypts PII fields and persists them. Writes EMPLOYEE_PII_UPDATED
// audit log after a successful update.
func (uc *SensitiveUseCase) UpdateSensitive(
	ctx context.Context,
	employeeID uuid.UUID,
	req UpdateSensitiveRequest,
	callerID uuid.UUID,
	callerRoles []string,
	ip string,
) error {
	if !hasSensitivePermission(callerRoles) {
		return domain.ErrInsufficientPermission
	}

	// Verify employee exists before encrypting/writing.
	if _, err := uc.repo.FindByID(ctx, employeeID); err != nil {
		return fmt.Errorf("sensitive.UpdateSensitive: %w", err)
	}

	encryptField := func(plain *string, fieldName string) (*string, error) {
		if plain == nil {
			return nil, nil
		}
		enc, err := uc.cipher.Encrypt(*plain)
		if err != nil {
			return nil, fmt.Errorf("encrypt %s: %w", fieldName, err)
		}
		return &enc, nil
	}

	parseDate := func(s *string) *time.Time {
		if s == nil {
			return nil
		}
		t, err := time.Parse("2006-01-02", *s)
		if err != nil {
			return nil
		}
		return &t
	}

	p := domain.UpdateSensitiveParams{
		ID:              employeeID,
		CccdIssuedDate:  parseDate(req.CCCDIssuedDate),
		CccdIssuedPlace: req.CCCDIssuedPlace,
		PassportNumber:  req.PassportNumber,
		PassportExpiry:  parseDate(req.PassportExpiry),
		BankName:        req.BankName,
		BankBranch:      req.BankBranch,
		UpdatedBy:       &callerID,
	}

	var err error
	if p.CccdEncrypted, err = encryptField(req.CCCD, "cccd"); err != nil {
		return fmt.Errorf("sensitive.UpdateSensitive: %w", err)
	}
	if p.MstCaNhanEncrypted, err = encryptField(req.MSTCaNhan, "mst_ca_nhan"); err != nil {
		return fmt.Errorf("sensitive.UpdateSensitive: %w", err)
	}
	if p.SoBhxhEncrypted, err = encryptField(req.SoBHXH, "so_bhxh"); err != nil {
		return fmt.Errorf("sensitive.UpdateSensitive: %w", err)
	}
	if p.BankAccountEncrypted, err = encryptField(req.BankAccount, "bank_account"); err != nil {
		return fmt.Errorf("sensitive.UpdateSensitive: %w", err)
	}

	if err := uc.repo.UpdateSensitiveFields(ctx, p); err != nil {
		return fmt.Errorf("sensitive.UpdateSensitive: %w", err)
	}

	// Track which logical fields were updated for the audit metadata.
	updatedFields := []string{}
	if req.CCCD != nil {
		updatedFields = append(updatedFields, "cccd")
	}
	if req.MSTCaNhan != nil {
		updatedFields = append(updatedFields, "mst_ca_nhan")
	}
	if req.SoBHXH != nil {
		updatedFields = append(updatedFields, "so_bhxh")
	}
	if req.BankAccount != nil {
		updatedFields = append(updatedFields, "bank_account")
	}
	if req.CCCDIssuedDate != nil || req.CCCDIssuedPlace != nil {
		updatedFields = append(updatedFields, "cccd_metadata")
	}
	if req.PassportNumber != nil || req.PassportExpiry != nil {
		updatedFields = append(updatedFields, "passport")
	}
	if req.BankName != nil || req.BankBranch != nil {
		updatedFields = append(updatedFields, "bank_info")
	}

	if _, auditErr := uc.auditLog.Log(ctx, audit.Entry{
		UserID:     &callerID,
		Module:     "hrm",
		Resource:   "employees",
		ResourceID: &employeeID,
		Action:     "EMPLOYEE_PII_UPDATED",
		NewValue:   map[string]any{"fields_updated": updatedFields},
		IPAddress:  ip,
	}); auditErr != nil {
		log.Printf("ERROR sensitive.UpdateSensitive audit EMPLOYEE_PII_UPDATED employee=%s caller=%s: %v",
			employeeID, callerID, auditErr)
	}

	return nil
}
