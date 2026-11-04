package orders

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusFulfilled Status = "fulfilled"
	StatusCancelled Status = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid            PaymentStatus = "unpaid"
	PaymentStatusProcessing        PaymentStatus = "processing"
	PaymentStatusPaid              PaymentStatus = "paid"
	PaymentStatusFailed            PaymentStatus = "failed"
	PaymentStatusPartiallyRefunded PaymentStatus = "partially_refunded"
	PaymentStatusRefunded          PaymentStatus = "refunded"
)

type Order struct {
	ID                      string
	CustomerID              *string
	CustomerEmail           string
	OrderStatus             Status
	PaymentStatus           PaymentStatus
	Currency                string
	SubtotalAmount          int64
	DiscountAmount          int64
	TaxAmount               int64
	TotalAmount             int64
	AmountRefunded          int64
	StripeCheckoutSessionID *string
	StripePaymentIntentID   *string
	Items                   []OrderItem
	CreatedAt               time.Time
	UpdatedAt               time.Time
	PaidAt                  *time.Time
	FulfilledAt             *time.Time
	CancelledAt             *time.Time
}

type OrderItem struct {
	ID              string
	OrderID         string
	ProductID       string
	ProductName     string
	VariantName     *string
	SKU             *string
	Quantity        int
	UnitPriceAmount int64
	LineTotalAmount int64
	CreatedAt       time.Time
}

type CheckoutCustomerInput struct {
	Email     string
	FirstName string
	LastName  string
	Phone     string
}

type CreateCheckoutOrderInput struct {
	CartTokenHash string
	Customer      CheckoutCustomerInput
}
