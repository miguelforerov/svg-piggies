-- migrate:up

CREATE TABLE carts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id uuid REFERENCES customers (id) ON DELETE CASCADE,

    -- The browser receives the original random token. Only its cryptographic
    -- hash is stored so a database leak does not expose active guest carts.
    guest_token_hash text UNIQUE,

    status text NOT NULL DEFAULT 'active' CHECK (
        status IN ('active', 'converted', 'abandoned', 'expired')
    ),
    currency varchar(3) NOT NULL DEFAULT 'aud' CHECK (
        length(currency) = 3 AND currency = lower(currency)
    ),
    converted_order_id uuid UNIQUE
        REFERENCES orders (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL DEFAULT (now() + interval '30 days'),

    CONSTRAINT carts_owner_required CHECK (
        customer_id IS NOT NULL OR guest_token_hash IS NOT NULL
    ),
    CONSTRAINT carts_guest_token_hash_valid CHECK (
        guest_token_hash IS NULL OR guest_token_hash ~ '^[0-9a-f]{64}$'
    ),
    CONSTRAINT carts_conversion_valid CHECK (
        converted_order_id IS NULL OR status = 'converted'
    ),
    CONSTRAINT carts_expiration_valid CHECK (expires_at > created_at)
);

CREATE TABLE cart_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id uuid NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    quantity integer NOT NULL DEFAULT 1 CHECK (quantity BETWEEN 1 AND 99),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT cart_items_unique_product UNIQUE (cart_id, product_id)
);

-- A customer may have only one active cart. A guest cart is uniquely
-- identified by guest_token_hash.
CREATE UNIQUE INDEX carts_one_active_customer_idx
    ON carts (customer_id)
    WHERE customer_id IS NOT NULL AND status = 'active';

CREATE INDEX carts_status_expiration_idx
    ON carts (status, expires_at);

CREATE INDEX cart_items_product_idx
    ON cart_items (product_id);

CREATE FUNCTION set_carts_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER carts_set_updated_at
BEFORE UPDATE ON carts
FOR EACH ROW
EXECUTE FUNCTION set_carts_updated_at();

CREATE FUNCTION set_cart_items_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER cart_items_set_updated_at
BEFORE UPDATE ON cart_items
FOR EACH ROW
EXECUTE FUNCTION set_cart_items_updated_at();

-- Keep the cart timestamp current whenever its contents change. This supports
-- expiration, abandonment reporting, and most-recent-cart retrieval.
CREATE FUNCTION touch_cart_from_cart_items()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE carts SET updated_at = now() WHERE id = OLD.cart_id;
        RETURN OLD;
    END IF;

    IF TG_OP = 'UPDATE' AND OLD.cart_id <> NEW.cart_id THEN
        UPDATE carts SET updated_at = now()
        WHERE id IN (OLD.cart_id, NEW.cart_id);
    ELSE
        UPDATE carts SET updated_at = now() WHERE id = NEW.cart_id;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER cart_items_touch_cart
AFTER INSERT OR UPDATE OR DELETE ON cart_items
FOR EACH ROW
EXECUTE FUNCTION touch_cart_from_cart_items();

-- migrate:down

DROP TRIGGER cart_items_touch_cart ON cart_items;
DROP FUNCTION touch_cart_from_cart_items();
DROP TRIGGER cart_items_set_updated_at ON cart_items;
DROP FUNCTION set_cart_items_updated_at();
DROP TRIGGER carts_set_updated_at ON carts;
DROP FUNCTION set_carts_updated_at();
DROP TABLE cart_items;
DROP TABLE carts;
