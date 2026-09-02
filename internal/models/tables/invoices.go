package tables

import "time"

type Invoice struct {
	PublicID          string     `gorm:"column:public_id;primaryKey"`
	BillingAccountID  string     `gorm:"column:billing_account_id"`
	OrganizationID    *string    `gorm:"column:organization_id"`
	SubscriptionID    *string    `gorm:"column:subscription_id"`
	Provider          string     `gorm:"column:provider"`
	ProviderInvoiceID string     `gorm:"column:provider_invoice_id"`
	Number            *string    `gorm:"column:number"`
	Status            string     `gorm:"column:status"`
	Currency          string     `gorm:"column:currency"`
	TotalMinor        int        `gorm:"column:total_minor"`
	AmountDueMinor    int        `gorm:"column:amount_due_minor"`
	HostedInvoiceURL  *string    `gorm:"column:hosted_invoice_url"`
	InvoicePDFURL     *string    `gorm:"column:invoice_pdf_url"`
	PeriodStart       *time.Time `gorm:"column:period_start"`
	PeriodEnd         *time.Time `gorm:"column:period_end"`
	PaidAt            *time.Time `gorm:"column:paid_at"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime:false"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime:false"`
}

func (Invoice) TableName() string { return "invoices" }
