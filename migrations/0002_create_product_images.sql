-- migrate:up

CREATE TABLE product_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    r2_object_key text NOT NULL UNIQUE,
    original_filename text NOT NULL,
    content_type text NOT NULL,
    file_size_bytes bigint NOT NULL,
    width_px integer,
    height_px integer,
    alt_text text,
    display_order integer NOT NULL DEFAULT 0 CHECK (display_order >= 0),
    is_primary boolean NOT NULL DEFAULT false,
    r2_etag text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX product_images_product_order_idx
    ON product_images (product_id, display_order, id);

CREATE FUNCTION set_product_images_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER product_images_set_updated_at
BEFORE UPDATE ON product_images
FOR EACH ROW
EXECUTE FUNCTION set_product_images_updated_at();

-- migrate:down

DROP TRIGGER product_images_set_updated_at ON product_images;
DROP FUNCTION set_product_images_updated_at();
DROP TABLE product_images;
