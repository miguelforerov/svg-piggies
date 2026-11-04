package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
	"github.com/zdenaforero/svg-piggies/backend/internal/orders"
)

type Repository struct {
	provider database.Provider
}

func NewRepository(provider database.Provider) (*Repository, error) {
	if provider == nil {
		return nil, errors.New("database provider is required")
	}
	return &Repository{provider: provider}, nil
}

func (r *Repository) List(ctx context.Context) ([]orders.Order, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	rows, err := connection.Query(ctx, orderSelect+` ORDER BY o.created_at DESC, o.id`)
	if err != nil {
		return nil, mapError(err)
	}
	result := []orders.Order{}
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			rows.Close()
			return nil, mapError(err)
		}
		result = append(result, order)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, mapError(err)
	}
	rows.Close()

	for index := range result {
		items, err := loadItems(ctx, connection, result[index].ID)
		if err != nil {
			return nil, err
		}
		result[index].Items = items
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id string) (orders.Order, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return orders.Order{}, err
	}
	defer release(connection)
	return loadOrder(ctx, connection, id)
}

func (r *Repository) CreateFromCart(
	ctx context.Context,
	input orders.CreateCheckoutOrderInput,
) (orders.Order, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return orders.Order{}, err
	}
	defer release(connection)

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return orders.Order{}, fmt.Errorf("begin checkout transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(context.WithoutCancel(ctx)) }()

	var cartID, currency string
	err = transaction.QueryRow(ctx, `
		SELECT id::text, currency
		FROM carts
		WHERE guest_token_hash = $1
		  AND status = 'active'
		  AND expires_at > now()
		FOR UPDATE
	`, input.CartTokenHash).Scan(&cartID, &currency)
	if err != nil {
		return orders.Order{}, mapError(err)
	}

	type cartItem struct {
		productID       string
		productName     string
		productStatus   string
		quantity        int
		unitPriceAmount int64
		lineTotalAmount int64
	}
	itemRows, err := transaction.Query(ctx, `
		SELECT p.id::text, p.title, p.status::text, ci.quantity,
		       round(p.price * 100)::bigint AS unit_price_amount,
		       (round(p.price * 100)::bigint * ci.quantity)::bigint AS line_total_amount
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at, ci.id
	`, cartID)
	if err != nil {
		return orders.Order{}, mapError(err)
	}
	cartItems := []cartItem{}
	var subtotal int64
	for itemRows.Next() {
		var item cartItem
		if err := itemRows.Scan(
			&item.productID,
			&item.productName,
			&item.productStatus,
			&item.quantity,
			&item.unitPriceAmount,
			&item.lineTotalAmount,
		); err != nil {
			itemRows.Close()
			return orders.Order{}, mapError(err)
		}
		if item.productStatus != "active" {
			itemRows.Close()
			return orders.Order{}, orders.ErrProductUnavailable
		}
		subtotal += item.lineTotalAmount
		cartItems = append(cartItems, item)
	}
	if err := itemRows.Err(); err != nil {
		itemRows.Close()
		return orders.Order{}, mapError(err)
	}
	itemRows.Close()
	if len(cartItems) == 0 {
		return orders.Order{}, orders.ErrEmptyCart
	}

	var customerID string
	err = transaction.QueryRow(ctx, `
		INSERT INTO customers (email, first_name, last_name, phone)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''))
		ON CONFLICT (lower(email)) DO UPDATE
		SET first_name = COALESCE(EXCLUDED.first_name, customers.first_name),
		    last_name = COALESCE(EXCLUDED.last_name, customers.last_name),
		    phone = COALESCE(EXCLUDED.phone, customers.phone)
		RETURNING id::text
	`, input.Customer.Email, input.Customer.FirstName, input.Customer.LastName, input.Customer.Phone).Scan(&customerID)
	if err != nil {
		return orders.Order{}, mapError(err)
	}

	var orderID string
	err = transaction.QueryRow(ctx, `
		INSERT INTO orders (
			customer_id, customer_email, currency,
			subtotal_amount, total_amount
		)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id::text
	`, customerID, input.Customer.Email, currency, subtotal).Scan(&orderID)
	if err != nil {
		return orders.Order{}, mapError(err)
	}

	for _, item := range cartItems {
		_, err := transaction.Exec(ctx, `
			INSERT INTO order_items (
				order_id, product_id, product_name, quantity,
				unit_price_amount, line_total_amount
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, orderID, item.productID, item.productName, item.quantity,
			item.unitPriceAmount, item.lineTotalAmount)
		if err != nil {
			return orders.Order{}, mapError(err)
		}
	}

	result, err := transaction.Exec(ctx, `
		UPDATE carts
		SET customer_id = $2, status = 'converted', converted_order_id = $3
		WHERE id = $1 AND status = 'active'
	`, cartID, customerID, orderID)
	if err != nil {
		return orders.Order{}, mapError(err)
	}
	if result.RowsAffected() != 1 {
		return orders.Order{}, orders.ErrConflict
	}

	order, err := loadOrder(ctx, transaction, orderID)
	if err != nil {
		return orders.Order{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return orders.Order{}, fmt.Errorf("commit checkout transaction: %w", err)
	}
	return order, nil
}

func (r *Repository) UpdateStatus(
	ctx context.Context,
	id string,
	status orders.Status,
) (orders.Order, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return orders.Order{}, err
	}
	defer release(connection)

	result, err := connection.Exec(ctx, `
		UPDATE orders
		SET order_status = $2,
		    fulfilled_at = CASE
		        WHEN $2 = 'fulfilled' THEN COALESCE(fulfilled_at, now())
		        ELSE NULL
		    END,
		    cancelled_at = CASE
		        WHEN $2 = 'cancelled' THEN COALESCE(cancelled_at, now())
		        ELSE NULL
		    END
		WHERE id = $1
	`, id, status)
	if err != nil {
		return orders.Order{}, mapError(err)
	}
	if result.RowsAffected() == 0 {
		return orders.Order{}, orders.ErrNotFound
	}
	return loadOrder(ctx, connection, id)
}

const orderSelect = `
	SELECT o.id::text, o.customer_id::text, COALESCE(o.customer_email, ''),
	       o.order_status, o.payment_status, o.currency,
	       o.subtotal_amount, o.discount_amount, o.tax_amount,
	       o.total_amount, o.amount_refunded,
	       o.stripe_checkout_session_id, o.stripe_payment_intent_id,
	       o.created_at, o.updated_at, o.paid_at, o.fulfilled_at, o.cancelled_at
	FROM orders o
`

type scanner interface {
	Scan(dest ...any) error
}

func scanOrder(row scanner) (orders.Order, error) {
	var result orders.Order
	err := row.Scan(
		&result.ID,
		&result.CustomerID,
		&result.CustomerEmail,
		&result.OrderStatus,
		&result.PaymentStatus,
		&result.Currency,
		&result.SubtotalAmount,
		&result.DiscountAmount,
		&result.TaxAmount,
		&result.TotalAmount,
		&result.AmountRefunded,
		&result.StripeCheckoutSessionID,
		&result.StripePaymentIntentID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.PaidAt,
		&result.FulfilledAt,
		&result.CancelledAt,
	)
	return result, err
}

func loadOrder(ctx context.Context, querier database.Querier, id string) (orders.Order, error) {
	order, err := scanOrder(querier.QueryRow(ctx, orderSelect+` WHERE o.id = $1`, id))
	if err != nil {
		return orders.Order{}, mapError(err)
	}
	items, err := loadItems(ctx, querier, id)
	if err != nil {
		return orders.Order{}, err
	}
	order.Items = items
	return order, nil
}

func loadItems(
	ctx context.Context,
	querier database.Querier,
	orderID string,
) ([]orders.OrderItem, error) {
	rows, err := querier.Query(ctx, `
		SELECT id::text, order_id::text, product_id::text, product_name,
		       variant_name, sku, quantity, unit_price_amount,
		       line_total_amount, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at, id
	`, orderID)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	items := []orders.OrderItem{}
	for rows.Next() {
		var item orders.OrderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.VariantName,
			&item.SKU,
			&item.Quantity,
			&item.UnitPriceAmount,
			&item.LineTotalAmount,
			&item.CreatedAt,
		); err != nil {
			return nil, mapError(err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}
	return items, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return orders.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return orders.ErrConflict
	}
	return fmt.Errorf("orders database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
