-- migrate:up

CREATE TABLE product_files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    file_key text NOT NULL UNIQUE CHECK (btrim(file_key) <> ''),
    file_name text NOT NULL CHECK (btrim(file_name) <> ''),
    display_name text NOT NULL CHECK (btrim(display_name) <> ''),
    content_type text NOT NULL CHECK (btrim(content_type) <> ''),
    file_size_bytes bigint NOT NULL CHECK (file_size_bytes >= 0),
    r2_etag text,
    display_order integer NOT NULL DEFAULT 0 CHECK (display_order >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX product_files_product_order_idx
    ON product_files (product_id, display_order, id);

CREATE FUNCTION set_product_files_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER product_files_set_updated_at
BEFORE UPDATE ON product_files
FOR EACH ROW
EXECUTE FUNCTION set_product_files_updated_at();

-- migrate:down

DROP TRIGGER product_files_set_updated_at ON product_files;
DROP FUNCTION set_product_files_updated_at();
DROP TABLE product_files;