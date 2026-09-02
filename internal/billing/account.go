package billing

import "time"

const (
	OwnerUser         = "user"
	OwnerOrganization = "organization"
)

type BillingAccount struct {
	ID                 string
	OwnerType          string
	OwnerID            string
	Provider           string
	ProviderCustomerID *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (a *BillingAccount) Personal() bool {
	return a != nil && a.OwnerType == OwnerUser
}

func (a *BillingAccount) Organization() bool {
	return a != nil && a.OwnerType == OwnerOrganization
}

func (a *BillingAccount) CustomerID() string {
	if a == nil || a.ProviderCustomerID == nil {
		return ""
	}
	return *a.ProviderCustomerID
}
