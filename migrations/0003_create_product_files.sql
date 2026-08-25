-- migrate:up

CREATE TABLE product_files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    file_key text NOT NULL UNIQUE CHECK (btrim(file_key) <> ''),
    file_name text NOT NULL CHECK (btrim(file_name) <> ''),
    file_type text NOT NULL CHECK (btrim(file_type) <> ''),
    display_order integer NOT NULL DEFAULT 0 CHECK (display_order >= 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX product_files_product_order_idx
    ON product_files (product_id, display_order, id);

-- migrate:down

DROP TABLE product_files;
