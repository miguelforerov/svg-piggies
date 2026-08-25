-- migrate:up

CREATE TABLE product_relationships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    related_product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    display_order integer NOT NULL DEFAULT 0 CHECK (display_order >= 0),
    CONSTRAINT product_relationships_distinct_products
        CHECK (product_id <> related_product_id),
    CONSTRAINT product_relationships_unique_pair
        UNIQUE (product_id, related_product_id)
);

CREATE INDEX product_relationships_product_order_idx
    ON product_relationships (product_id, display_order, id);

CREATE INDEX product_relationships_related_product_idx
    ON product_relationships (related_product_id);

-- migrate:down

DROP TABLE product_relationships;
