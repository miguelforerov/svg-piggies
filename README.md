# SVG Piggies

Digital-product storefront built with StorePlate, a Go backend on Cloudflare
Workers, Neon PostgreSQL, Cloudflare R2, and Stripe.

## Development setup

The project requires Node.js 22.12 or later, Go 1.25 or later, npm, and Docker with
Compose. The expected runtime versions are recorded in `.nvmrc`, `.go-version`,
and `package.json`.

```sh
npm install
cp .env.example .env
cp .dev.vars.example .dev.vars
```

`.env` configures local tools, the Astro storefront, dbmate, and Hyperdrive's
local database connection. `.dev.vars` contains Worker runtime secrets and must
not be committed.

### Storefront

```sh
npm run dev
```

Astro serves the storefront at `http://localhost:4321` and reads the API origin
from `PUBLIC_API_BASE_URL`.

### Go API

Run the API as a normal Go HTTP server while developing business logic:

```sh
npm run backend:dev
```

The API listens on `http://localhost:8787`. Its public health endpoint is
`GET /api/health`. Run Go tests with:

```sh
npm run backend:test
```

### OpenAPI

The API contract is defined in `backend/api/openapi.yaml`. The project uses a
pinned `oapi-codegen` tool to generate strict `net/http` request, response, and
server types in `backend/api/generated`.

After changing the contract, regenerate the Go code:

```sh
npm run backend:generate
```

Generated files are committed. Do not edit `server.gen.go` manually. Implement
operations in `backend/internal/handlers` and keep business rules in the
feature service packages. The initial contract includes `GET /api/health` and the
collection CRUD routes rooted at `/api/admin/collections`.

Admin authentication is controlled by `AUTH_MODE`. Local development uses
`AUTH_MODE=development`, which bypasses authentication only when `APP_ENV` is
`development` or `test`. Deployed staging and production environments must use:

```dotenv
AUTH_MODE="cloudflare_access"
CLOUDFLARE_ACCESS_TEAM_DOMAIN="https://your-team.cloudflareaccess.com"
CLOUDFLARE_ACCESS_AUD="your-access-application-aud"
```

Cloudflare Access must be configured at the edge to protect both `/admin/*`
and `/api/admin/*`. The backend also validates the Access JWT from
`Cf-Access-Jwt-Assertion`, including its signature, issuer, audience, and
expiration. It refuses to start if development authentication is selected for
a non-local environment.

The native Go server reads `DATABASE_URL` from `.env`, validates that its
database name matches `DATABASE_NAME`, and keeps a small `pgx` connection pool.
The collection service is wired to the PostgreSQL repository at startup.

Run the Go API through the Cloudflare Workers WebAssembly adapter with:

```sh
npm run worker:dev
```

`wrangler.jsonc` declares the runtime contract:

- `HYPERDRIVE` for PostgreSQL connectivity;
- `PRODUCT_FILES` for private R2 objects;
- committed non-secret application variables; and
- required Stripe secrets loaded locally from `.dev.vars`.

In the Worker, the same repositories acquire a connection through the
`HYPERDRIVE` binding for each database operation. Hyperdrive performs the
connection pooling, so the Worker does not create another long-lived pool.

Before deploying, replace the placeholder Hyperdrive ID and provision the R2
buckets. Store production Stripe values with `wrangler secret put`, never in
`wrangler.jsonc`.

Build and deploy commands are:

```sh
npm run worker:build
npm run worker:deploy
```

The deploy command selects Wrangler's `production` environment. Before using
it, replace the production Hyperdrive ID and storefront origin in
`wrangler.jsonc`. The Hyperdrive configuration must target `svg_piggies`.

## Database migrations

This project uses [dbmate](https://github.com/amacneil/dbmate) as a lightweight,
SQL-first PostgreSQL migration tool. Dbmate is a development/CI tool; it does
not run inside the Cloudflare Worker.

### Setup

Local database development requires Docker with Compose, Node.js, and npm.
Install the project dependencies:

```sh
npm install
```

Copy `.env.example` to `.env`, then set `DATABASE_URL` to the development
database connection string:

```sh
cp .env.example .env
```

Create the local PostgreSQL database and apply all migrations:

```sh
npm run db:create
```

This starts the `postgres` service from `compose.yaml`, waits until PostgreSQL
accepts connections, and applies every pending migration to both the
development and test databases. The database data is kept in the named
`postgres_data` Docker volume between container restarts.

The database names are fixed by environment:

- production: `svg_piggies`;
- development: `svg_piggies_development`;
- test: `svg_piggies_test`.

The test database is created by the PostgreSQL initialization script only when
the Docker data volume is first created. If this project already has an older
local volume, create `svg_piggies_test` manually or recreate that development
volume when it is safe to discard its local data.

Stop the local database without deleting its data:

```sh
npm run db:stop
```

Start an existing local database again with:

```sh
npm run db:start
```

The local defaults are development-only. Change `POSTGRES_PASSWORD` and the
matching password in `DATABASE_URL` if the database is exposed beyond the local
machine.

For Neon, use a direct (non-pooled) connection string with TLS, rather than the
Hyperdrive or pooled application connection string:

```dotenv
DATABASE_URL="postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require"
```

For production, `DATABASE` must be `svg_piggies`. Dbmate should still receive a
direct Neon URL; the application Worker receives its runtime URL from
Hyperdrive instead.

After setting the Neon development branch URL, run `npm run db:migrate`. Do not
run `db:create` for Neon because the Neon database is created through Neon and
does not use the local Compose service.

The committed `.env.example` configures dbmate to:

- read migrations from `./migrations`;
- reject migrations that would be applied out of order; and
- skip automatic `schema.sql` dumps, keeping the migration files as the schema
  history for now.

Never commit `.env` or database credentials.

### Commands

Create a timestamped migration:

```sh
npm run db:migrate:new -- add_orders
```

Apply all pending migrations:

```sh
npm run db:migrate
```

Apply migrations to the local test database:

```sh
npm run db:migrate:test
```

Roll back the most recently applied migration:

```sh
npm run db:migrate:rollback
```

Show applied and pending migrations:

```sh
npm run db:migrate:status
```

### Migration format

Keep migrations as explicit SQL files with both sections present:

```sql
-- migrate:up

CREATE TABLE example (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid()
);

-- migrate:down

DROP TABLE example;
```

Dbmate wraps each Up or Down section in a transaction by default, so migration
files should not include their own `BEGIN;` or `COMMIT;`. Only opt out for a
PostgreSQL operation that cannot run in a transaction by changing the relevant
header to `-- migrate:up transaction:false` or
`-- migrate:down transaction:false`.

Write the Down section in reverse dependency order so rolling back does not
violate foreign-key dependencies. Once a migration has been applied to a shared
environment, add a new migration instead of editing the applied file.
