package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zdenaforero/svg-piggies/backend/internal/carts"
	"github.com/zdenaforero/svg-piggies/backend/internal/database"
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

func (r *Repository) CreateGuest(ctx context.Context, tokenHash string) (carts.Cart, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return carts.Cart{}, err
	}
	defer release(connection)

	var cart carts.Cart
	err = connection.QueryRow(ctx, `
		INSERT INTO carts (guest_token_hash)
		VALUES ($1)
		RETURNING id::text, status, currency, created_at, updated_at, expires_at
	`, tokenHash).Scan(
		&cart.ID,
		&cart.Status,
		&cart.Currency,
		&cart.CreatedAt,
		&cart.UpdatedAt,
		&cart.ExpiresAt,
	)
	if err != nil {
		return carts.Cart{}, mapError(err)
	}
	cart.Items = []carts.CartItem{}
	return cart, nil
}

func (r *Repository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (carts.Cart, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return carts.Cart{}, err
	}
	defer release(connection)

	return loadCartByTokenHash(ctx, connection, tokenHash)
}

func (r *Repository) AddItem(
	ctx context.Context,
	tokenHash string,
	productID string,
	quantity int,
) (carts.Cart, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return carts.Cart{}, err
	}
	defer release(connection)

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return carts.Cart{}, fmt.Errorf("begin cart transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(context.WithoutCancel(ctx)) }()

	cartID, err := lockActiveCart(ctx, transaction, tokenHash)
	if err != nil {
		return carts.Cart{}, err
	}
	if err := requireActiveProduct(ctx, transaction, productID); err != nil {
		return carts.Cart{}, err
	}

	result, err := transaction.Exec(ctx, `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id) DO UPDATE
		SET quantity = cart_items.quantity + EXCLUDED.quantity
		WHERE cart_items.quantity + EXCLUDED.quantity <= $4
	`, cartID, productID, quantity, carts.MaxQuantity)
	if err != nil {
		return carts.Cart{}, mapError(err)
	}
	if result.RowsAffected() == 0 {
		return carts.Cart{}, fmt.Errorf(
			"%w: resulting quantity cannot exceed %d",
			carts.ErrInvalidInput,
			carts.MaxQuantity,
		)
	}

	cart, err := loadCartByID(ctx, transaction, cartID)
	if err != nil {
		return carts.Cart{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return carts.Cart{}, fmt.Errorf("commit cart transaction: %w", err)
	}
	return cart, nil
}

func (r *Repository) UpdateItem(
	ctx context.Context,
	tokenHash string,
	productID string,
	quantity int,
) (carts.Cart, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return carts.Cart{}, err
	}
	defer release(connection)

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return carts.Cart{}, fmt.Errorf("begin cart transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(context.WithoutCancel(ctx)) }()

	cartID, err := lockActiveCart(ctx, transaction, tokenHash)
	if err != nil {
		return carts.Cart{}, err
	}
	if err := requireActiveProduct(ctx, transaction, productID); err != nil {
		return carts.Cart{}, err
	}

	result, err := transaction.Exec(ctx, `
		UPDATE cart_items
		SET quantity = $1
		WHERE cart_id = $2 AND product_id = $3
	`, quantity, cartID, productID)
	if err != nil {
		return carts.Cart{}, mapError(err)
	}
	if result.RowsAffected() == 0 {
		return carts.Cart{}, carts.ErrNotFound
	}

	cart, err := loadCartByID(ctx, transaction, cartID)
	if err != nil {
		return carts.Cart{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return carts.Cart{}, fmt.Errorf("commit cart transaction: %w", err)
	}
	return cart, nil
}

func (r *Repository) DeleteItem(
	ctx context.Context,
	tokenHash string,
	productID string,
) (carts.Cart, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return carts.Cart{}, err
	}
	defer release(connection)

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return carts.Cart{}, fmt.Errorf("begin cart transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(context.WithoutCancel(ctx)) }()

	cartID, err := lockActiveCart(ctx, transaction, tokenHash)
	if err != nil {
		return carts.Cart{}, err
	}
	result, err := transaction.Exec(ctx, `
		DELETE FROM cart_items
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID)
	if err != nil {
		return carts.Cart{}, mapError(err)
	}
	if result.RowsAffected() == 0 {
		return carts.Cart{}, carts.ErrNotFound
	}

	cart, err := loadCartByID(ctx, transaction, cartID)
	if err != nil {
		return carts.Cart{}, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return carts.Cart{}, fmt.Errorf("commit cart transaction: %w", err)
	}
	return cart, nil
}

func (r *Repository) Clear(ctx context.Context, tokenHash string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	transaction, err := connection.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin cart transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(context.WithoutCancel(ctx)) }()

	cartID, err := lockActiveCart(ctx, transaction, tokenHash)
	if err != nil {
		return err
	}
	if _, err := transaction.Exec(ctx, `
		DELETE FROM cart_items WHERE cart_id = $1
	`, cartID); err != nil {
		return mapError(err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit cart transaction: %w", err)
	}
	return nil
}

func lockActiveCart(
	ctx context.Context,
	querier database.Querier,
	tokenHash string,
) (string, error) {
	var cartID string
	err := querier.QueryRow(ctx, `
		SELECT id::text
		FROM carts
		WHERE guest_token_hash = $1
		  AND status = 'active'
		  AND expires_at > now()
		FOR UPDATE
	`, tokenHash).Scan(&cartID)
	if err != nil {
		return "", mapError(err)
	}
	return cartID, nil
}

func requireActiveProduct(
	ctx context.Context,
	querier database.Querier,
	productID string,
) error {
	var exists bool
	if err := querier.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM products
			WHERE id = $1 AND status = 'active'
		)
	`, productID).Scan(&exists); err != nil {
		return mapError(err)
	}
	if !exists {
		return carts.ErrProductUnavailable
	}
	return nil
}

func loadCartByTokenHash(
	ctx context.Context,
	querier database.Querier,
	tokenHash string,
) (carts.Cart, error) {
	return loadCart(ctx, querier, `
		SELECT id::text, status, currency, created_at, updated_at, expires_at
		FROM carts
		WHERE guest_token_hash = $1
		  AND status = 'active'
		  AND expires_at > now()
	`, tokenHash)
}

func loadCartByID(
	ctx context.Context,
	querier database.Querier,
	cartID string,
) (carts.Cart, error) {
	return loadCart(ctx, querier, `
		SELECT id::text, status, currency, created_at, updated_at, expires_at
		FROM carts
		WHERE id = $1
	`, cartID)
}

func loadCart(
	ctx context.Context,
	querier database.Querier,
	query string,
	argument string,
) (carts.Cart, error) {
	var cart carts.Cart
	if err := querier.QueryRow(ctx, query, argument).Scan(
		&cart.ID,
		&cart.Status,
		&cart.Currency,
		&cart.CreatedAt,
		&cart.UpdatedAt,
		&cart.ExpiresAt,
	); err != nil {
		return carts.Cart{}, mapError(err)
	}

	rows, err := querier.Query(ctx, `
		SELECT
			ci.id::text,
			ci.product_id::text,
			p.title,
			p.slug,
			p.price::text,
			ci.quantity,
			ci.created_at,
			ci.updated_at
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at, ci.id
	`, cart.ID)
	if err != nil {
		return carts.Cart{}, mapError(err)
	}
	defer rows.Close()

	cart.Items = []carts.CartItem{}
	for rows.Next() {
		var item carts.CartItem
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Name,
			&item.Slug,
			&item.Price,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return carts.Cart{}, mapError(err)
		}
		cart.Items = append(cart.Items, item)
	}
	if err := rows.Err(); err != nil {
		return carts.Cart{}, mapError(err)
	}
	return cart, nil
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return carts.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return carts.ErrConflict
		case "23503":
			return carts.ErrProductUnavailable
		case "22P02":
			return carts.ErrInvalidInput
		}
	}
	return fmt.Errorf("carts database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
