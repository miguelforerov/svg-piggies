BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '60s';

DO $guard$
BEGIN
    IF current_database() <> 'svg_piggies_test' THEN
        RAISE EXCEPTION 'Wrong database: %', current_database();
    END IF;
    IF (SELECT count(*) FROM public.schema_migrations WHERE version = '0003') <> 1 THEN
        RAISE EXCEPTION 'Migration 0003 must occur exactly once';
    END IF;
    IF to_regclass('public.product_files') IS NULL THEN
        RAISE EXCEPTION 'product_files is missing';
    END IF;
END
$guard$;

LOCK TABLE public.product_files IN SHARE ROW EXCLUSIVE MODE;
LOCK TABLE public.schema_migrations IN SHARE MODE;

DO $guard$
BEGIN
    IF (SELECT count(*) FROM public.schema_migrations WHERE version = '0003') <> 1 THEN
        RAISE EXCEPTION 'Migration 0003 changed during synchronization';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_class
        WHERE oid = 'public.product_files'::regclass AND relkind = 'r'
          AND relpersistence = 'p' AND NOT relrowsecurity AND NOT relforcerowsecurity
    ) THEN
        RAISE EXCEPTION 'Unexpected product_files table properties';
    END IF;
END
$guard$;

CREATE TABLE IF NOT EXISTS public.product_file_consumptions (
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

LOCK TABLE public.product_file_consumptions IN SHARE ROW EXCLUSIVE MODE;

-- Build expected catalog objects on this PostgreSQL server. This avoids relying
-- on manually predicted pg_get_constraintdef/pg_get_indexdef formatting.
CREATE TEMP TABLE expected_product_files (
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
    CONSTRAINT product_files_file_key_key UNIQUE (file_key),
    CONSTRAINT product_files_file_key_check CHECK (btrim(file_key) <> ''),
    CONSTRAINT product_files_file_name_check CHECK (btrim(file_name) <> ''),
    CONSTRAINT product_files_display_name_check CHECK (btrim(display_name) <> ''),
    CONSTRAINT product_files_format_check CHECK (format IN ('svg', 'png', 'dxf', 'pdf')),
    CONSTRAINT product_files_content_type_check CHECK (btrim(content_type) <> ''),
    CONSTRAINT product_files_file_size_bytes_check CHECK (file_size_bytes >= 0),
    CONSTRAINT product_files_availability_status_check
        CHECK (availability_status IN ('active', 'retired', 'blocked')),
    CONSTRAINT product_files_confirmation_fingerprint_check
        CHECK (confirmation_fingerprint ~ '^[0-9a-f]{64}$'),
    CONSTRAINT product_files_fingerprint_key_version_check CHECK (fingerprint_key_version > 0),
    CONSTRAINT product_files_display_order_check CHECK (display_order >= 0)
) ON COMMIT DROP;
CREATE INDEX product_files_product_order_idx
    ON expected_product_files (product_id, display_order, id);

CREATE TEMP TABLE expected_product_file_consumptions (
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
) ON COMMIT DROP;
CREATE INDEX product_file_consumptions_cleanup_idx
    ON expected_product_file_consumptions (expires_at, file_key);

-- IF NOT EXISTS alone would accept an unrelated or malformed index. Create only
-- when absent; the complete catalog comparison below validates either path.
DO $index$
BEGIN
    IF to_regclass('public.product_file_consumptions_cleanup_idx') IS NULL THEN
        CREATE INDEX product_file_consumptions_cleanup_idx
            ON public.product_file_consumptions (expires_at, file_key);
    END IF;
END
$index$;

DO $guard$
DECLARE
    actual_table regclass;
    expected_table regclass;
    table_name text;
    actual_value jsonb;
    expected_value jsonb;
BEGIN
    IF (SELECT count(*) FROM public.schema_migrations WHERE version = '0003') <> 1 THEN
        RAISE EXCEPTION 'Migration 0003 changed during synchronization';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_class
        WHERE oid = 'public.product_file_consumptions'::regclass AND relkind = 'r'
          AND relpersistence = 'p' AND NOT relrowsecurity AND NOT relforcerowsecurity
    ) THEN
        RAISE EXCEPTION 'Unexpected product_file_consumptions table properties';
    END IF;

    FOR actual_table, expected_table, table_name IN
        SELECT 'public.product_files'::regclass, 'pg_temp.expected_product_files'::regclass, 'product_files'
        UNION ALL
        SELECT 'public.product_file_consumptions'::regclass,
               'pg_temp.expected_product_file_consumptions'::regclass,
               'product_file_consumptions'
    LOOP
        SELECT COALESCE(jsonb_object_agg(a.attname, jsonb_build_array(
                   format_type(a.atttypid, a.atttypmod), a.attnotnull,
                   pg_get_expr(d.adbin, d.adrelid), a.attgenerated, a.attidentity,
                   a.attcollation::regcollation::text, a.attstorage)), '{}'::jsonb)
          INTO actual_value
          FROM pg_attribute a
          LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum
         WHERE a.attrelid=actual_table AND a.attnum>0 AND NOT a.attisdropped;
        SELECT COALESCE(jsonb_object_agg(a.attname, jsonb_build_array(
                   format_type(a.atttypid, a.atttypmod), a.attnotnull,
                   pg_get_expr(d.adbin, d.adrelid), a.attgenerated, a.attidentity,
                   a.attcollation::regcollation::text, a.attstorage)), '{}'::jsonb)
          INTO expected_value
          FROM pg_attribute a
          LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum
         WHERE a.attrelid=expected_table AND a.attnum>0 AND NOT a.attisdropped;
        IF actual_value IS DISTINCT FROM expected_value THEN
            RAISE EXCEPTION 'Unexpected % columns, defaults, or nullability', table_name;
        END IF;

        SELECT COALESCE(jsonb_object_agg(conname, jsonb_build_array(
                   contype, pg_get_constraintdef(oid), convalidated,
                   condeferrable, condeferred)), '{}'::jsonb)
          INTO actual_value FROM pg_constraint
         WHERE conrelid=actual_table
           AND (table_name <> 'product_files' OR contype <> 'f');
        SELECT COALESCE(jsonb_object_agg(conname, jsonb_build_array(
                   contype, pg_get_constraintdef(oid), convalidated,
                   condeferrable, condeferred)), '{}'::jsonb)
          INTO expected_value FROM pg_constraint
         WHERE conrelid=expected_table;
        IF actual_value IS DISTINCT FROM expected_value THEN
            RAISE EXCEPTION 'Unexpected % PRIMARY KEY, UNIQUE, or CHECK definitions', table_name;
        END IF;

        SELECT COALESCE(jsonb_object_agg(ic.relname, jsonb_build_array(
                   am.amname, i.indisunique, i.indisprimary, i.indnullsnotdistinct,
                   i.indisvalid, i.indisready, i.indnkeyatts, i.indnatts,
                   i.indoption::text, i.indclass::text, i.indcollation::text,
                   pg_get_expr(i.indpred, i.indrelid),
                   pg_get_expr(i.indexprs, i.indrelid),
                   (SELECT jsonb_agg(pg_get_indexdef(i.indexrelid, g.pos, true) ORDER BY g.pos)
                      FROM generate_series(1, i.indnatts) AS g(pos)))), '{}'::jsonb)
          INTO actual_value
          FROM pg_index i
          JOIN pg_class ic ON ic.oid=i.indexrelid
          JOIN pg_am am ON am.oid=ic.relam
         WHERE i.indrelid=actual_table;
        SELECT COALESCE(jsonb_object_agg(ic.relname, jsonb_build_array(
                   am.amname, i.indisunique, i.indisprimary, i.indnullsnotdistinct,
                   i.indisvalid, i.indisready, i.indnkeyatts, i.indnatts,
                   i.indoption::text, i.indclass::text, i.indcollation::text,
                   pg_get_expr(i.indpred, i.indrelid),
                   pg_get_expr(i.indexprs, i.indrelid),
                   (SELECT jsonb_agg(pg_get_indexdef(i.indexrelid, g.pos, true) ORDER BY g.pos)
                      FROM generate_series(1, i.indnatts) AS g(pos)))), '{}'::jsonb)
          INTO expected_value
          FROM pg_index i
          JOIN pg_class ic ON ic.oid=i.indexrelid
          JOIN pg_am am ON am.oid=ic.relam
         WHERE i.indrelid=expected_table;
        IF actual_value IS DISTINCT FROM expected_value THEN
            RAISE EXCEPTION 'Unexpected % index columns, order, uniqueness, or predicates', table_name;
        END IF;
    END LOOP;

    IF (SELECT count(*) FROM pg_constraint c
         WHERE c.conrelid='public.product_files'::regclass AND c.contype='f') <> 1
       OR NOT EXISTS (
        SELECT 1 FROM pg_constraint c
         WHERE c.conrelid='public.product_files'::regclass
           AND c.conname='product_files_product_id_fkey'
           AND c.contype='f' AND c.confrelid='public.products'::regclass
           AND c.conkey=ARRAY[(SELECT attnum FROM pg_attribute
                               WHERE attrelid='public.product_files'::regclass
                                 AND attname='product_id')]::smallint[]
           AND c.confkey=ARRAY[(SELECT attnum FROM pg_attribute
                                WHERE attrelid='public.products'::regclass
                                  AND attname='id')]::smallint[]
           AND c.confdeltype='c' AND c.confupdtype='a' AND c.confmatchtype='s'
           AND NOT c.condeferrable AND NOT c.condeferred AND c.convalidated
       ) THEN
        RAISE EXCEPTION 'Unexpected product_files product foreign key';
    END IF;

    IF (SELECT count(*) FROM pg_trigger t
         WHERE t.tgrelid='public.product_files'::regclass AND NOT t.tgisinternal) <> 1
       OR NOT EXISTS (
        SELECT 1 FROM pg_trigger t
         WHERE t.tgrelid='public.product_files'::regclass AND NOT t.tgisinternal
           AND t.tgname='product_files_set_updated_at'
           AND t.tgfoid=to_regprocedure('public.set_product_files_updated_at()')
           AND t.tgtype=19 AND t.tgenabled='O' AND t.tgqual IS NULL
           AND t.tgnargs=0 AND t.tgattr::text=''
           AND t.tgconstraint=0
       )
       OR NOT EXISTS (
        SELECT 1 FROM pg_proc p
        JOIN pg_namespace n ON n.oid=p.pronamespace
        JOIN pg_language l ON l.oid=p.prolang
        WHERE n.nspname='public' AND p.proname='set_product_files_updated_at'
          AND pg_get_function_identity_arguments(p.oid)=''
          AND pg_get_function_result(p.oid)='trigger'
          AND l.lanname='plpgsql'
          AND NOT p.prosecdef AND NOT p.proisstrict
          AND p.provolatile='v' AND p.proparallel='u'
          AND p.proconfig IS NULL
          AND regexp_replace(p.prosrc, '^[[:space:]]+|[[:space:]]+$', '', 'g')
              = E'BEGIN\n    NEW.updated_at = now();\n    RETURN NEW;\nEND;'
       ) THEN
        RAISE EXCEPTION 'Unexpected product_files update function or trigger';
    END IF;
END
$guard$;

-- A snapshot may be older than mutable display metadata, but its immutable
-- identity and file attributes must still match the original file.
DO $snapshot$
DECLARE
    snapshot_row record;
BEGIN
    IF EXISTS (
        SELECT 1 FROM public.product_file_consumptions c
         WHERE (c.response_snapshot - 'id' - 'productId' - 'fileName' - 'displayName'
                 - 'format' - 'contentType' - 'fileSizeBytes' - 'availabilityStatus'
                 - 'displayOrder' - 'createdAt' - 'updatedAt') <> '{}'::jsonb
            OR jsonb_typeof(c.response_snapshot->'id') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'productId') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'fileName') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'displayName') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'format') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'contentType') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'fileSizeBytes') IS DISTINCT FROM 'number'
            OR jsonb_typeof(c.response_snapshot->'availabilityStatus') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'displayOrder') IS DISTINCT FROM 'number'
            OR jsonb_typeof(c.response_snapshot->'createdAt') IS DISTINCT FROM 'string'
            OR jsonb_typeof(c.response_snapshot->'updatedAt') IS DISTINCT FROM 'string'
            OR c.response_snapshot->>'id' IS DISTINCT FROM c.product_file_id::text
            OR c.response_snapshot->>'productId' IS DISTINCT FROM c.product_id::text
            OR c.response_snapshot->>'format' NOT IN ('svg','png','dxf','pdf')
            OR c.response_snapshot->>'availabilityStatus' NOT IN ('active','retired','blocked')
            OR c.response_snapshot->>'fileSizeBytes' !~ '^[0-9]+$'
            OR c.response_snapshot->>'displayOrder' !~ '^[0-9]+$'
            OR btrim(c.response_snapshot->>'fileName') = ''
            OR btrim(c.response_snapshot->>'displayName') = ''
            OR btrim(c.response_snapshot->>'contentType') = ''
            OR (c.token_digest IS NULL) <> (c.expires_at IS NULL)
    ) THEN
        RAISE EXCEPTION 'Malformed or contradictory historical response snapshot';
    END IF;
    FOR snapshot_row IN
        SELECT file_key, response_snapshot FROM public.product_file_consumptions
    LOOP
        BEGIN
            PERFORM (snapshot_row.response_snapshot->>'createdAt')::timestamptz;
            PERFORM (snapshot_row.response_snapshot->>'updatedAt')::timestamptz;
        EXCEPTION WHEN invalid_datetime_format OR datetime_field_overflow THEN
            RAISE EXCEPTION 'Invalid historical timestamp for file_key %', snapshot_row.file_key;
        END;
    END LOOP;
    IF EXISTS (
        SELECT 1 FROM public.product_files f
        JOIN public.product_file_consumptions c USING (file_key)
        WHERE c.product_id <> f.product_id
           OR c.product_file_id <> f.id
           OR c.confirmation_fingerprint <> f.confirmation_fingerprint
           OR c.fingerprint_key_version <> f.fingerprint_key_version
           OR c.response_snapshot->>'id' IS DISTINCT FROM f.id::text
           OR c.response_snapshot->>'productId' IS DISTINCT FROM f.product_id::text
           OR c.response_snapshot->>'fileName' IS DISTINCT FROM f.file_name
           OR c.response_snapshot->>'format' IS DISTINCT FROM f.format
           OR c.response_snapshot->>'contentType' IS DISTINCT FROM f.content_type
           OR c.response_snapshot->>'fileSizeBytes' IS DISTINCT FROM f.file_size_bytes::text
           OR (c.response_snapshot->>'createdAt')::timestamptz IS DISTINCT FROM f.created_at
           OR c.deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Existing consumption contradicts its original product_files row';
    END IF;
END
$snapshot$;

INSERT INTO public.product_file_consumptions
    (file_key, product_id, product_file_id, confirmation_fingerprint,
     fingerprint_key_version, response_snapshot)
SELECT f.file_key, f.product_id, f.id, f.confirmation_fingerprint,
       f.fingerprint_key_version,
       jsonb_build_object(
         'id', f.id::text, 'productId', f.product_id::text,
         'fileName', f.file_name, 'displayName', f.display_name,
         'format', f.format, 'contentType', f.content_type,
         'fileSizeBytes', f.file_size_bytes,
         'availabilityStatus', f.availability_status,
         'displayOrder', f.display_order,
         'createdAt', f.created_at, 'updatedAt', f.updated_at)
FROM public.product_files f
ON CONFLICT (file_key) DO NOTHING;

DO $guard$
BEGIN
    IF EXISTS (
        SELECT 1 FROM public.product_files f
        LEFT JOIN public.product_file_consumptions c USING (file_key)
        WHERE c.file_key IS NULL OR c.product_file_id <> f.id
           OR c.product_id <> f.product_id
    ) THEN
        RAISE EXCEPTION 'Consumption backfill incomplete';
    END IF;
END
$guard$;

COMMIT;
