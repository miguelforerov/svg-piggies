package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/zdenaforero/svg-piggies/backend/internal/customers"
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

func (r *Repository) List(ctx context.Context) ([]customers.Customer, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release(connection)

	rows, err := connection.Query(ctx, `
		SELECT id::text, email, first_name, last_name, phone,
		       stripe_customer_id, auth_user_id, marketing_consent,
		       marketing_consent_at, created_at, updated_at
		FROM customers
		ORDER BY created_at DESC, id
	`)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	result := []customers.Customer{}
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return nil, mapError(err)
		}
		result = append(result, customer)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err)
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id string) (customers.Customer, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return customers.Customer{}, err
	}
	defer release(connection)
	return getCustomer(ctx, connection, id)
}

func (r *Repository) Create(
	ctx context.Context,
	input customers.CreateCustomerInput,
) (customers.Customer, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return customers.Customer{}, err
	}
	defer release(connection)

	customer, err := scanCustomer(connection.QueryRow(ctx, `
		INSERT INTO customers (
			email, first_name, last_name, phone,
			marketing_consent, marketing_consent_at
		)
		VALUES (
			$1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), $5,
			CASE WHEN $5 THEN now() ELSE NULL END
		)
		RETURNING id::text, email, first_name, last_name, phone,
		          stripe_customer_id, auth_user_id, marketing_consent,
		          marketing_consent_at, created_at, updated_at
	`, input.Email, input.FirstName, input.LastName, input.Phone, input.MarketingConsent))
	if err != nil {
		return customers.Customer{}, mapError(err)
	}
	return customer, nil
}

func (r *Repository) Update(
	ctx context.Context,
	id string,
	input customers.UpdateCustomerInput,
) (customers.Customer, error) {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return customers.Customer{}, err
	}
	defer release(connection)

	customer, err := scanCustomer(connection.QueryRow(ctx, `
		UPDATE customers
		SET email = $2,
		    first_name = NULLIF($3, ''),
		    last_name = NULLIF($4, ''),
		    phone = NULLIF($5, ''),
		    marketing_consent = $6,
		    marketing_consent_at = CASE
		        WHEN $6 AND NOT marketing_consent THEN now()
		        WHEN NOT $6 THEN NULL
		        ELSE marketing_consent_at
		    END
		WHERE id = $1
		RETURNING id::text, email, first_name, last_name, phone,
		          stripe_customer_id, auth_user_id, marketing_consent,
		          marketing_consent_at, created_at, updated_at
	`, id, input.Email, input.FirstName, input.LastName, input.Phone, input.MarketingConsent))
	if err != nil {
		return customers.Customer{}, mapError(err)
	}
	return customer, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	connection, err := r.provider.Acquire(ctx)
	if err != nil {
		return err
	}
	defer release(connection)

	result, err := connection.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if result.RowsAffected() == 0 {
		return customers.ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func getCustomer(
	ctx context.Context,
	querier database.Querier,
	id string,
) (customers.Customer, error) {
	customer, err := scanCustomer(querier.QueryRow(ctx, `
		SELECT id::text, email, first_name, last_name, phone,
		       stripe_customer_id, auth_user_id, marketing_consent,
		       marketing_consent_at, created_at, updated_at
		FROM customers
		WHERE id = $1
	`, id))
	if err != nil {
		return customers.Customer{}, mapError(err)
	}
	return customer, nil
}

func scanCustomer(row scanner) (customers.Customer, error) {
	var result customers.Customer
	err := row.Scan(
		&result.ID,
		&result.Email,
		&result.FirstName,
		&result.LastName,
		&result.Phone,
		&result.StripeCustomerID,
		&result.AuthUserID,
		&result.MarketingConsent,
		&result.MarketingConsentAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	return result, err
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
		return customers.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return customers.ErrConflict
	}
	return fmt.Errorf("customers database operation: %w", err)
}

func release(connection *database.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = connection.Release(ctx)
}
