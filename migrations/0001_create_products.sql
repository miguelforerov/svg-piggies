-- migrate:up

CREATE TYPE product_status AS ENUM (
    'draft',
    'active',
    'archived'
);

CREATE TABLE products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title text NOT NULL CHECK (btrim(title) <> ''),
    slug text NOT NULL UNIQUE CHECK (btrim(slug) <> ''),
    description text NOT NULL DEFAULT '',
    price numeric(12, 2) NOT NULL CHECK (price >= 0),
    status product_status NOT NULL DEFAULT 'draft',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE FUNCTION set_products_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER products_set_updated_at
BEFORE UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION set_products_updated_at();

-- migrate:down

DROP TRIGGER products_set_updated_at ON products;
DROP FUNCTION set_products_updated_at();
DROP TABLE products;
DROP TYPE product_status;
