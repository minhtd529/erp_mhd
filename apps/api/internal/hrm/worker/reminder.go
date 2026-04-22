// Package worker provides the HRM daily reminder Asynq job.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	notifDomain "github.com/mdh/erp-audit/api/internal/notification/domain"
	hrmDomain "github.com/mdh/erp-audit/api/internal/hrm/domain"
)

// TaskHRMReminder is the Asynq task type for the daily HRM alert scan.
const TaskHRMReminder = "hrm:daily-reminder"

// HRMReminderUseCase performs the daily scan and writes inbox notifications.
type HRMReminderUseCase struct {
	notifRepo    notifDomain.Repository
	certRepo     hrmDomain.CertificationRepository
	recordRepo   hrmDomain.TrainingRecordRepository
	contractRepo hrmDomain.ContractRepository
	provRepo     hrmDomain.ProvisioningRepository
}

// NewHRMReminderUseCase wires the daily reminder scan.
func NewHRMReminderUseCase(
	notifRepo notifDomain.Repository,
	certRepo hrmDomain.CertificationRepository,
	recordRepo hrmDomain.TrainingRecordRepository,
	contractRepo hrmDomain.ContractRepository,
	provRepo hrmDomain.ProvisioningRepository,
) *HRMReminderUseCase {
	return &HRMReminderUseCase{
		notifRepo:    notifRepo,
		certRepo:     certRepo,
		recordRepo:   recordRepo,
		contractRepo: contractRepo,
		provRepo:     provRepo,
	}
}

// NewHRMReminderHandler returns an Asynq handler that runs the daily HRM scan.
func NewHRMReminderHandler(uc *HRMReminderUseCase) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		return uc.Run(ctx)
	}
}

// Run executes all four alert scans. Individual scan errors are logged but do
// not stop subsequent scans so a partial run is still useful.
func (uc *HRMReminderUseCase) Run(ctx context.Context) error {
	var errs []error
	if err := uc.certExpiryAlerts(ctx); err != nil {
		errs = append(errs, fmt.Errorf("cert-expiry: %w", err))
	}
	if err := uc.cpeDeadlineAlerts(ctx); err != nil {
		errs = append(errs, fmt.Errorf("cpe-deadline: %w", err))
	}
	if err := uc.provisioningExpiredAlerts(ctx); err != nil {
		errs = append(errs, fmt.Errorf("provisioning-expired: %w", err))
	}
	if err := uc.contractExpiryAlerts(ctx); err != nil {
		errs = append(errs, fmt.Errorf("contract-expiry: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("hrm-reminder: %v", errs)
	}
	return nil
}

// certExpiryAlerts notifies employees when their certification expires in 90, 60,
// or 30 days. The source_ref encodes the bucket so each threshold fires exactly once.
func (uc *HRMReminderUseCase) certExpiryAlerts(ctx context.Context) error {
	alerts, err := uc.certRepo.ListExpiringAlerts(ctx, 90)
	if err != nil {
		return err
	}
	for _, a := range alerts {
		days := int(time.Until(a.ExpiryDate).Hours() / 24)
		bucket := certBucket(days)
		if bucket == 0 {
			continue
		}
		if err := uc.notifRepo.Insert(ctx, notifDomain.InsertParams{
			UserID:    a.UserID,
			Type:      notifDomain.TypeCertExpiry,
			Title:     "Chứng chỉ sắp hết hạn",
			Body:      fmt.Sprintf("%s sẽ hết hạn trong %d ngày", a.CertName, days),
			SourceRef: fmt.Sprintf("cert:%s:%d", a.CertID, bucket),
		}); err != nil {
			return err
		}
	}
	return nil
}

// certBucket maps the remaining days to the notification threshold (90/60/30).
// Returns 0 if the cert is already past-due (should not happen with the query filter).
func certBucket(days int) int {
	switch {
	case days <= 0:
		return 0
	case days <= 30:
		return 30
	case days <= 60:
		return 60
	default:
		return 90
	}
}

// cpeDeadlineAlerts notifies employees who are behind on CPE hours for the
// current year. Fires once per employee per year (source_ref dedup).
func (uc *HRMReminderUseCase) cpeDeadlineAlerts(ctx context.Context) error {
	year := time.Now().Year()
	deficits, err := uc.recordRepo.ListCPEDeficit(ctx, year)
	if err != nil {
		return err
	}
	for _, d := range deficits {
		missing := d.RequiredHours - d.TotalHours
		if err := uc.notifRepo.Insert(ctx, notifDomain.InsertParams{
			UserID: d.UserID,
			Type:   notifDomain.TypeCPEDeadline,
			Title:  "Cần hoàn thành CPE trước cuối năm",
			Body: fmt.Sprintf(
				"Bạn cần hoàn thành thêm %.0f giờ CPE trước ngày 31/12/%d",
				missing, year,
			),
			SourceRef: fmt.Sprintf("cpe:%s:%d", d.EmployeeID, year),
		}); err != nil {
			return err
		}
	}
	return nil
}

// provisioningExpiredAlerts notifies the requester when a PENDING provisioning
// request has passed its expiry date.
func (uc *HRMReminderUseCase) provisioningExpiredAlerts(ctx context.Context) error {
	expired, err := uc.provRepo.ListExpiredPending(ctx)
	if err != nil {
		return err
	}
	for _, e := range expired {
		if err := uc.notifRepo.Insert(ctx, notifDomain.InsertParams{
			UserID: e.RequestedBy,
			Type:   notifDomain.TypeProvisioningExpired,
			Title:  "Yêu cầu cấp quyền đã hết hạn",
			Body:   "Yêu cầu cấp quyền tài khoản đã hết hạn, vui lòng tạo yêu cầu mới nếu cần",
			SourceRef: fmt.Sprintf("prov:%s", e.RequestID),
		}); err != nil {
			return err
		}
	}
	return nil
}

// contractExpiryAlerts notifies employees when their current contract ends within
// 30 days. Fires once per contract (source_ref dedup).
func (uc *HRMReminderUseCase) contractExpiryAlerts(ctx context.Context) error {
	alerts, err := uc.contractRepo.ListExpiringContracts(ctx, 30)
	if err != nil {
		return err
	}
	for _, a := range alerts {
		if err := uc.notifRepo.Insert(ctx, notifDomain.InsertParams{
			UserID: a.UserID,
			Type:   notifDomain.TypeContractExpiry,
			Title:  "Hợp đồng sắp hết hạn",
			Body:   fmt.Sprintf("Hợp đồng lao động sẽ hết hạn vào %s", a.EndDate.Format("02/01/2006")),
			SourceRef: fmt.Sprintf("contract:%s", a.ContractID),
		}); err != nil {
			return err
		}
	}
	return nil
}
