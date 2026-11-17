-- migrate:up

CREATE TABLE public.product_files (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL,
    file_key text NOT NULL,
    file_name text NOT NULL,
    display_name text NOT NULL,
    format text NOT NULL,
    content_type text NOT NULL,
    file_size_bytes bigint NOT NULL,
    r2_etag text,
    availability_status text NOT NULL DEFAULT 'active',
    confirmation_fingerprint text NOT NULL,
    fingerprint_key_version integer NOT NULL,
    display_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT product_files_pkey PRIMARY KEY (id),
    CONSTRAINT product_files_product_id_fkey
        FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE,
    CONSTRAINT product_files_file_key_key UNIQUE (file_key),
    CONSTRAINT product_files_file_key_check CHECK (btrim(file_key) <> ''),
    CONSTRAINT product_files_file_name_check CHECK (btrim(file_name) <> ''),
    CONSTRAINT product_files_display_name_check CHECK (btrim(display_name) <> ''),
    CONSTRAINT product_files_format_check
        CHECK (format IN ('svg', 'png', 'dxf', 'pdf')),
    CONSTRAINT product_files_content_type_check CHECK (btrim(content_type) <> ''),
    CONSTRAINT product_files_file_size_bytes_check CHECK (file_size_bytes >= 0),
    CONSTRAINT product_files_availability_status_check
        CHECK (availability_status IN ('active', 'retired', 'blocked')),
    CONSTRAINT product_files_confirmation_fingerprint_check
        CHECK (confirmation_fingerprint ~ '^[0-9a-f]{64}$'),
    CONSTRAINT product_files_fingerprint_key_version_check
        CHECK (fingerprint_key_version > 0),
    CONSTRAINT product_files_display_order_check CHECK (display_order >= 0)
);

CREATE INDEX product_files_product_order_idx
    ON public.product_files (product_id, display_order, id);

-- This ledger is intentionally independent of products and product_files: deleting
-- either visible row must never release a successfully consumed upload key.
CREATE TABLE public.product_file_consumptions (
    file_key text NOT NULL,
    product_id uuid NOT NULL,
    product_file_id uuid NOT NULL,
    token_digest text,
    confirmation_fingerprint text NOT NULL,
    fingerprint_key_version integer NOT NULL,
    expires_at timestamptz,
    consumed_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    response_snapshot jsonb NOT NULL,
    CONSTRAINT product_file_consumptions_pkey PRIMARY KEY (file_key),
    CONSTRAINT product_file_consumptions_product_file_id_key UNIQUE (product_file_id),
    CONSTRAINT product_file_consumptions_token_digest_key UNIQUE (token_digest),
    CONSTRAINT product_file_consumptions_file_key_check CHECK (btrim(file_key) <> ''),
    CONSTRAINT product_file_consumptions_token_digest_check
        CHECK (token_digest IS NULL OR token_digest ~ '^[0-9a-f]{64}$'),
    CONSTRAINT product_file_consumptions_fingerprint_check
        CHECK (confirmation_fingerprint ~ '^[0-9a-f]{64}$'),
    CONSTRAINT product_file_consumptions_key_version_check CHECK (fingerprint_key_version > 0),
    CONSTRAINT product_file_consumptions_snapshot_check
        CHECK (jsonb_typeof(response_snapshot) = 'object')
);

CREATE INDEX product_file_consumptions_cleanup_idx
    ON public.product_file_consumptions (expires_at, file_key);

CREATE FUNCTION public.set_product_files_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER product_files_set_updated_at
BEFORE UPDATE ON public.product_files
FOR EACH ROW
EXECUTE FUNCTION public.set_product_files_updated_at();

-- migrate:down

DROP TABLE public.product_file_consumptions;
DROP TRIGGER product_files_set_updated_at ON public.product_files;
DROP FUNCTION public.set_product_files_updated_at();
DROP TABLE public.product_files;
