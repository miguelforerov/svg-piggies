-- The development database is created by POSTGRES_DB. Create the separate
-- test database when Docker initializes a fresh PostgreSQL data volume.
SELECT 'CREATE DATABASE svg_piggies_test'
WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'svg_piggies_test'
)\gexec
