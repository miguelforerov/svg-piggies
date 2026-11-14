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

DROP TRIGGER product_files_set_updated_at ON public.product_files;
DROP FUNCTION public.set_product_files_updated_at();
DROP TABLE public.product_files;
