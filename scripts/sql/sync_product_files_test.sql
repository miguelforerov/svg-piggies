BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

DO $target$
BEGIN
    IF current_database() <> 'svg_piggies_test' THEN
        RAISE EXCEPTION 'Wrong database: %', current_database();
    END IF;
END
$target$;

LOCK TABLE public.product_files IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.schema_migrations IN SHARE MODE;

DO $guard$
DECLARE
    v_columns jsonb;
    v_constraints jsonb;
    v_indexes jsonb;
    v_triggers jsonb;
    v_file_type_attnum smallint;
    v_file_type_constraint_oid oid;
BEGIN
    IF (SELECT count(*) FROM public.schema_migrations WHERE version = '0003') <> 1 THEN
        RAISE EXCEPTION 'Migration 0003 must occur exactly once';
    END IF;
    IF EXISTS (SELECT 1 FROM public.product_files) THEN
        RAISE EXCEPTION 'product_files is not empty';
    END IF;

    SELECT jsonb_object_agg(column_name, jsonb_build_array(data_type, is_nullable, column_default))
      INTO v_columns
      FROM information_schema.columns
     WHERE table_schema = 'public' AND table_name = 'product_files';
    IF v_columns IS DISTINCT FROM $expected${
      "id":["uuid","NO","gen_random_uuid()"],
      "product_id":["uuid","NO",null],
      "file_key":["text","NO",null],
      "file_name":["text","NO",null],
      "file_type":["text","NO",null],
      "display_order":["integer","NO","0"],
      "created_at":["timestamp with time zone","NO","now()"]
    }$expected$::jsonb THEN
        RAISE EXCEPTION 'Unexpected test columns: %', v_columns;
    END IF;

    SELECT jsonb_object_agg(conname, pg_get_constraintdef(oid))
      INTO v_constraints
      FROM pg_constraint
     WHERE conrelid = 'public.product_files'::regclass;
    IF v_constraints IS DISTINCT FROM $expected${
      "product_files_display_order_check":"CHECK ((display_order >= 0))",
      "product_files_file_key_check":"CHECK ((btrim(file_key) <> ''::text))",
      "product_files_file_key_key":"UNIQUE (file_key)",
      "product_files_file_name_check":"CHECK ((btrim(file_name) <> ''::text))",
      "product_files_file_type_check":"CHECK ((btrim(file_type) <> ''::text))",
      "product_files_pkey":"PRIMARY KEY (id)",
      "product_files_product_id_fkey":"FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE"
    }$expected$::jsonb THEN
        RAISE EXCEPTION 'Unexpected test constraints: %', v_constraints;
    END IF;

    SELECT jsonb_object_agg(indexname, indexdef)
      INTO v_indexes
      FROM pg_indexes
     WHERE schemaname = 'public' AND tablename = 'product_files';
    IF v_indexes IS DISTINCT FROM $expected${
      "product_files_file_key_key":"CREATE UNIQUE INDEX product_files_file_key_key ON public.product_files USING btree (file_key)",
      "product_files_pkey":"CREATE UNIQUE INDEX product_files_pkey ON public.product_files USING btree (id)",
      "product_files_product_order_idx":"CREATE INDEX product_files_product_order_idx ON public.product_files USING btree (product_id, display_order, id)"
    }$expected$::jsonb THEN
        RAISE EXCEPTION 'Unexpected test indexes: %', v_indexes;
    END IF;

    SELECT COALESCE(jsonb_object_agg(tgname, pg_get_triggerdef(oid)), '{}'::jsonb)
      INTO v_triggers
      FROM pg_trigger
     WHERE tgrelid = 'public.product_files'::regclass AND NOT tgisinternal;
    IF v_triggers <> '{}'::jsonb
       OR to_regprocedure('public.set_product_files_updated_at()') IS NOT NULL THEN
        RAISE EXCEPTION 'Unexpected test function or trigger';
    END IF;

    SELECT attnum INTO v_file_type_attnum
      FROM pg_attribute
     WHERE attrelid = 'public.product_files'::regclass
       AND attname = 'file_type' AND NOT attisdropped;
    SELECT oid INTO v_file_type_constraint_oid
      FROM pg_constraint
     WHERE conrelid = 'public.product_files'::regclass
       AND conname = 'product_files_file_type_check';
    IF v_file_type_attnum IS NULL OR v_file_type_constraint_oid IS NULL THEN
        RAISE EXCEPTION 'file_type or its verified constraint is absent';
    END IF;
    IF (SELECT count(*)
          FROM pg_depend d
         WHERE d.refclassid = 'pg_class'::regclass
           AND d.refobjid = 'public.product_files'::regclass
           AND d.refobjsubid = v_file_type_attnum) <> 2
       OR (SELECT count(DISTINCT d.deptype)
             FROM pg_depend d
            WHERE d.refclassid = 'pg_class'::regclass
              AND d.refobjid = 'public.product_files'::regclass
              AND d.refobjsubid = v_file_type_attnum
              AND d.classid = 'pg_constraint'::regclass
              AND d.objid = v_file_type_constraint_oid
              AND d.deptype IN ('a', 'n')) <> 2 THEN
        RAISE EXCEPTION 'Unexpected dependencies on file_type';
    END IF;
END
$guard$;

ALTER TABLE public.product_files
    ADD COLUMN display_name text NOT NULL,
    ADD COLUMN format text NOT NULL,
    ADD COLUMN content_type text NOT NULL,
    ADD COLUMN file_size_bytes bigint NOT NULL,
    ADD COLUMN r2_etag text,
    ADD COLUMN availability_status text NOT NULL DEFAULT 'active',
    ADD COLUMN confirmation_fingerprint text NOT NULL,
    ADD COLUMN fingerprint_key_version integer NOT NULL,
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

ALTER TABLE public.product_files
    ADD CONSTRAINT product_files_display_name_check CHECK (btrim(display_name) <> ''),
    ADD CONSTRAINT product_files_format_check CHECK (format IN ('svg', 'png', 'dxf', 'pdf')),
    ADD CONSTRAINT product_files_content_type_check CHECK (btrim(content_type) <> ''),
    ADD CONSTRAINT product_files_file_size_bytes_check CHECK (file_size_bytes >= 0),
    ADD CONSTRAINT product_files_availability_status_check CHECK (availability_status IN ('active', 'retired', 'blocked')),
    ADD CONSTRAINT product_files_confirmation_fingerprint_check CHECK (confirmation_fingerprint ~ '^[0-9a-f]{64}$'),
    ADD CONSTRAINT product_files_fingerprint_key_version_check CHECK (fingerprint_key_version > 0);

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

ALTER TABLE public.product_files DROP CONSTRAINT product_files_file_type_check;
ALTER TABLE public.product_files DROP COLUMN file_type RESTRICT;

COMMIT;
