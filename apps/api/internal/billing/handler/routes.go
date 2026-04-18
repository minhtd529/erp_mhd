package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires all billing routes under /api/v1.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	invoiceH *InvoiceHandler,
	paymentH *PaymentHandler,
	memoH *MemoHandler,
	arH *ARHandler,
	reportH *ReportHandler,
	authMW gin.HandlerFunc,
) {
	requirePartner := middleware.RequireRole("FIRM_PARTNER")
	requireManager := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER")
	requireStaff := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF")

	invoices := v1.Group("/invoices", authMW)
	{
		invoices.GET("", requireStaff, invoiceH.List)
		invoices.GET("/approval-queue", requireManager, invoiceH.ApprovalQueue)
		invoices.POST("", requireManager, invoiceH.Create)
		invoices.POST("/generate-from-engagement", requireManager, invoiceH.GenerateFromEngagement)
		invoices.GET("/:id", requireStaff, invoiceH.GetByID)
		invoices.PUT("/:id", requirePartner, invoiceH.Update)
		invoices.DELETE("/:id", requirePartner, invoiceH.Delete)

		// State transitions
		invoices.POST("/:id/send", requirePartner, invoiceH.Send)
		invoices.POST("/:id/confirm", requirePartner, invoiceH.Confirm)
		invoices.POST("/:id/issue", requirePartner, invoiceH.Issue)
		invoices.POST("/:id/cancel", requirePartner, invoiceH.Cancel)

		// Line items
		invoices.GET("/:id/line-items", requireStaff, invoiceH.ListLineItems)
		invoices.POST("/:id/line-items", requirePartner, invoiceH.AddLineItem)
		invoices.DELETE("/:id/line-items/:item_id", requirePartner, invoiceH.DeleteLineItem)

		// Payments nested under invoice
		invoices.GET("/:id/payments", requireStaff, paymentH.ListByInvoice)
		invoices.POST("/:id/record-payment", requirePartner, paymentH.Record)

		// Credit notes
		invoices.POST("/:id/credit-note", requirePartner, memoH.Create)
	}

	// Standalone payment operations
	payments := v1.Group("/payments", authMW)
	{
		payments.PUT("/:id", requirePartner, paymentH.Update)
		payments.POST("/:id/clear", requirePartner, paymentH.Clear)
		payments.POST("/:id/dispute", requirePartner, paymentH.Dispute)
		payments.DELETE("/:id", requirePartner, paymentH.Reverse)
	}

	// Credit notes list
	creditNotes := v1.Group("/credit-notes", authMW)
	{
		creditNotes.GET("", requireStaff, memoH.List)
	}

	// Accounts Receivable
	ar := v1.Group("/ar", authMW)
	{
		ar.GET("/aging", requirePartner, arH.Aging)
		ar.GET("/outstanding", requirePartner, arH.Outstanding)
	}

	// Billing reports
	reports := v1.Group("/billing/reports", authMW)
	{
		reports.GET("/period-summary", requireManager, reportH.PeriodSummary)
		reports.GET("/payment-summary", requireManager, reportH.PaymentSummary)
		// Export period summary as XLSX: GET /billing/reports/period-summary/export?format=xlsx&start=&end=
		reports.GET("/period-summary/export", requireManager, reportH.ExportPeriodSummary)
	}

	// Invoice export (CSV or XLSX via ?format= param)
	invoiceExport := v1.Group("/invoices", authMW)
	{
		invoiceExport.GET("/export", requireManager, reportH.ExportInvoices)
	}
}
