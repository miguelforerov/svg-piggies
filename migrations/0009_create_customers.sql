-- migrate:up

CREATE TABLE customers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL CHECK (btrim(email) <> ''),
    first_name text,
    last_name text,
    phone text,
    stripe_customer_id text UNIQUE,
    auth_user_id text UNIQUE,
    marketing_consent boolean NOT NULL DEFAULT false,
    marketing_consent_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX customers_email_unique_idx
    ON customers (lower(email));

CREATE FUNCTION set_customers_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER customers_set_updated_at
BEFORE UPDATE ON customers
FOR EACH ROW
EXECUTE FUNCTION set_customers_updated_at();

-- migrate:down

DROP TRIGGER customers_set_updated_at ON customers;
DROP FUNCTION set_customers_updated_at();
DROP TABLE customers;
