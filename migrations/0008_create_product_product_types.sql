-- migrate:up

CREATE TABLE product_product_types (
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    product_type_id uuid NOT NULL REFERENCES product_types (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, product_type_id)
);

CREATE INDEX product_product_types_product_type_idx
    ON product_product_types (product_type_id, product_id);

-- migrate:down

DROP TABLE product_product_types;
