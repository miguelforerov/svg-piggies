-- migrate:up

CREATE TABLE product_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    url text NOT NULL CHECK (btrim(url) <> ''),
    display_order integer NOT NULL DEFAULT 0 CHECK (display_order >= 0)
);

CREATE INDEX product_images_product_order_idx
    ON product_images (product_id, display_order, id);

-- migrate:down

DROP TABLE product_images;
