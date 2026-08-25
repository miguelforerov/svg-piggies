-- migrate:up

CREATE TABLE product_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (btrim(name) <> ''),
    slug text NOT NULL UNIQUE CHECK (btrim(slug) <> ''),
    description text NOT NULL DEFAULT ''
);

-- migrate:down

DROP TABLE product_types;
