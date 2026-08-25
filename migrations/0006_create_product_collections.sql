-- migrate:up

CREATE TABLE product_collections (
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    collection_id uuid NOT NULL REFERENCES collections (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, collection_id)
);

CREATE INDEX product_collections_collection_idx
    ON product_collections (collection_id, product_id);

-- migrate:down

DROP TABLE product_collections;
