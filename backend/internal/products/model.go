package products

import "time"

type Status string

const (
	StatusDraft    Status = "draft"
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

type Product struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       string    `json:"price"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateProductInput struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      Status `json:"status"`
}

type UpdateProductInput struct {
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      Status `json:"status"`
}
