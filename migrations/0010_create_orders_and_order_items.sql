-- migrate:up

CREATE TABLE orders (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id uuid REFERENCES customers (id) ON DELETE SET NULL,
    customer_email text,
    order_status text NOT NULL DEFAULT 'pending' CHECK (
        order_status IN ('pending', 'fulfilled', 'cancelled')
    ),
    payment_status text NOT NULL DEFAULT 'unpaid' CHECK (
        payment_status IN (
            'unpaid',
            'processing',
            'paid',
            'failed',
            'partially_refunded',
            'refunded'
        )
    ),
    currency varchar(3) NOT NULL DEFAULT 'aud' CHECK (
        length(currency) = 3 AND currency = lower(currency)
    ),
    subtotal_amount bigint NOT NULL CHECK (subtotal_amount >= 0),
    discount_amount bigint NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    tax_amount bigint NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    total_amount bigint NOT NULL CHECK (total_amount >= 0),
    amount_refunded bigint NOT NULL DEFAULT 0 CHECK (
        amount_refunded >= 0 AND amount_refunded <= total_amount
    ),
    stripe_checkout_session_id text UNIQUE,
    stripe_payment_intent_id text UNIQUE,
    stripe_customer_id text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    paid_at timestamptz,
    fulfilled_at timestamptz,
    cancelled_at timestamptz
);

CREATE TABLE order_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id uuid NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    product_name text NOT NULL CHECK (btrim(product_name) <> ''),
    variant_name text,
    sku text,
    quantity integer NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price_amount bigint NOT NULL CHECK (unit_price_amount >= 0),
    line_total_amount bigint NOT NULL CHECK (line_total_amount >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX orders_customer_created_at_idx
    ON orders (customer_id, created_at DESC)
    WHERE customer_id IS NOT NULL;

CREATE INDEX orders_status_created_at_idx
    ON orders (payment_status, order_status, created_at);

CREATE INDEX order_items_order_id_idx
    ON order_items (order_id);

CREATE INDEX order_items_product_order_idx
    ON order_items (product_id, order_id);

CREATE FUNCTION set_orders_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER orders_set_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION set_orders_updated_at();

-- migrate:down

DROP TRIGGER orders_set_updated_at ON orders;
DROP FUNCTION set_orders_updated_at();
DROP TABLE order_items;
DROP TABLE orders;
