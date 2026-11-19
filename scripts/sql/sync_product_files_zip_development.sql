BEGIN;

DO $guard$
BEGIN
    IF current_database() <> 'svg_piggies_development' THEN
        RAISE EXCEPTION 'Unexpected database for product file ZIP synchronization';
    END IF;
    IF (SELECT count(*) FROM public.schema_migrations WHERE version = '0003') <> 1 THEN
        RAISE EXCEPTION 'Migration 0003 must be registered exactly once';
    END IF;
    IF to_regclass('public.product_files') IS NULL THEN
        RAISE EXCEPTION 'product_files table is missing';
    END IF;
END
$guard$;

LOCK TABLE public.product_files IN ACCESS EXCLUSIVE MODE;

-- PostgreSQL produces both expressions below, so the comparison does not
-- depend on a manually assumed pg_get_constraintdef rendering.
CREATE TEMP TABLE product_file_format_expected (
    format text,
    CONSTRAINT expected_old CHECK (format IN ('svg', 'png', 'dxf', 'pdf')),
    CONSTRAINT expected_new CHECK (format IN ('svg', 'png', 'dxf', 'pdf', 'zip'))
) ON COMMIT DROP;

DO $sync$
DECLARE
    actual record;
    old_definition text;
    new_definition text;
BEGIN
    SELECT pg_get_constraintdef(oid) INTO old_definition
      FROM pg_constraint
     WHERE conrelid = 'pg_temp.product_file_format_expected'::regclass
       AND conname = 'expected_old';
    SELECT pg_get_constraintdef(oid) INTO new_definition
      FROM pg_constraint
     WHERE conrelid = 'pg_temp.product_file_format_expected'::regclass
       AND conname = 'expected_new';
    SELECT contype, convalidated, conislocal, coninhcount, connoinherit,
           pg_get_constraintdef(oid) AS definition
      INTO actual
      FROM pg_constraint
     WHERE conrelid = 'public.product_files'::regclass
       AND conname = 'product_files_format_check';
    IF NOT FOUND OR actual.contype <> 'c' OR NOT actual.convalidated
       OR NOT actual.conislocal OR actual.coninhcount <> 0 OR actual.connoinherit THEN
        RAISE EXCEPTION 'Unexpected product_files format constraint metadata';
    END IF;
    IF EXISTS (SELECT 1 FROM public.product_files
               WHERE format NOT IN ('svg', 'png', 'dxf', 'pdf', 'zip')) THEN
        RAISE EXCEPTION 'Unexpected product_files format data';
    END IF;
    IF actual.definition = old_definition THEN
        ALTER TABLE public.product_files
            DROP CONSTRAINT product_files_format_check,
            ADD CONSTRAINT product_files_format_check
                CHECK (format IN ('svg', 'png', 'dxf', 'pdf', 'zip'));
    ELSIF actual.definition IS DISTINCT FROM new_definition THEN
        RAISE EXCEPTION 'Unexpected product_files format constraint definition';
    END IF;
END
$sync$;

COMMIT;
