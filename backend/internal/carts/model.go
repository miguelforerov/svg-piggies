package carts

import "time"

const (
	StatusActive = "active"
	MaxQuantity  = 99
)

type Cart struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Currency  string     `json:"currency"`
	Items     []CartItem `json:"items"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
}

type CartItem struct {
	ID        string    `json:"id"`
	ProductID string    `json:"productId"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Price     string    `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreatedCart struct {
	Token string `json:"token"`
	Cart  Cart   `json:"cart"`
}

type AddItemInput struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
