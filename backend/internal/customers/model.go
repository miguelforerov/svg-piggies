package customers

import "time"

type Customer struct {
	ID                 string
	Email              string
	FirstName          *string
	LastName           *string
	Phone              *string
	StripeCustomerID   *string
	AuthUserID         *string
	MarketingConsent   bool
	MarketingConsentAt *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateCustomerInput struct {
	Email            string
	FirstName        string
	LastName         string
	Phone            string
	MarketingConsent bool
}

type UpdateCustomerInput = CreateCustomerInput
