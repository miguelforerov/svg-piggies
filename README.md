# SVG Piggies

Digital-product storefront built with Astro and shadcn/ui, a Go backend on Cloudflare
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

Admin authentication is controlled by `AUTH_MODE`. The native local server can
use `AUTH_MODE=development` with `APP_ENV=development` or `test`. A Worker refuses
to start unless `APP_ENV=production`, `AUTH_MODE=cloudflare_access`, and an
administrator allowlist are configured. Deployed environments must use:

```dotenv
AUTH_MODE="cloudflare_access"
CLOUDFLARE_ACCESS_TEAM_DOMAIN="https://your-team.cloudflareaccess.com"
CLOUDFLARE_ACCESS_AUD="your-access-application-aud"
ADMIN_EMAILS="admin@example.invalid,second-admin@example.invalid"
```

Cloudflare Access must be configured at the edge to protect both `/admin/*`
and `/api/admin/*`. The backend also validates the Access JWT from
`Cf-Access-Jwt-Assertion`, including its signature, issuer, audience, and
expiration. It refuses to start if development authentication is selected for
a non-local environment. The backend reads the `email` claim only after JWT
verification, normalizes it, and permits admin routes only when that email is
in the comma-separated `ADMIN_EMAILS` list. A missing or empty list blocks
startup; an authenticated user outside the list receives 403.

The base Wrangler configuration disables `workers_dev`. `npm run worker:deploy`
runs `worker:check` before Wrangler; it checks the production environment,
Access settings, Hyperdrive ID, bucket names, and required secret names. The
check cannot verify secret values, and a direct `wrangler deploy` bypasses the
script. There is currently no CI deployment workflow in this repository; any
future production deployment job must run `npm run worker:check` as a required
gate. The Worker also rejects development authentication at startup. The local
native server remains available through `npm run backend:dev`; Worker previews
require the production authentication configuration.

The native Go server reads `DATABASE_URL` from `.env`, validates that its
database name matches `DATABASE_NAME`, and keeps a small `pgx` connection pool.
The collection service is wired to the PostgreSQL repository at startup.

Run the Go API through the Cloudflare Workers WebAssembly adapter with:

```sh
npm run worker:dev
```

`backend/wrangler.jsonc` declares the API Worker's runtime contract:

- `HYPERDRIVE` for PostgreSQL connectivity;
- `PRODUCT_FILES` for private R2 downloads and `PRODUCT_ASSETS` for product images;
- committed non-secret application variables; and
- required secrets loaded locally from `backend/.dev.vars`.

Production secret names are `ADMIN_EMAILS`, `STRIPE_SECRET_KEY`,
`STRIPE_WEBHOOK_SECRET`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, and
`R2_ENDPOINT`. Configure their values outside the repository; the deployment
checker verifies only that their names are declared.

### R2 storage prerequisites

The S3 client uses `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_ENDPOINT`
(`https://<account-id>.r2.cloudflarestorage.com`), `R2_REGION=auto`, and the
two `PRODUCT_*_BUCKET` names. The native server reads these values from `.env`.
The Worker reads its bucket names and region from `backend/wrangler.jsonc` and
its credentials and endpoint from Worker secrets (locally,
`backend/.dev.vars`, using `backend/.dev.vars.example` as a template). R2
bindings and S3 environment variables are separate; keep their bucket names
aligned. Development uses `svgpiggies-files-dev` and `svgpiggies-assets-dev`;
production uses `svgpiggies-files` and `svgpiggies-assets`. Keep the files
bucket private and never expose the secret access key to the storefront.

Before testing browser uploads from the admin panel, configure a CORS policy
on **both** buckets with the storefront origin, `PUT`, and every header sent
with the signed request, including `Content-Type` when present. Add `GET`,
`HEAD`, or other methods only when the final browser flow needs them. The
browser must send the exact signed header values returned with the PUT URL.
The API must authorize each operation and derive the bucket and object key from
trusted server-side data. A presigned PUT URL is reusable until it expires;
verify uploaded metadata, including size, with HEAD afterward. The client has
not yet been exercised against real R2 or with HEAD/DELETE in a Worker runtime.

The admin image flow starts with `POST /api/admin/products/{productId}/images/uploads`
and finishes with `POST /api/admin/products/{productId}/images`. The first
response contains a short-lived signed upload token, a PUT URL, and the exact
headers to send (`Content-Type` and `If-None-Match: *`). Include both headers in
the assets bucket CORS `AllowedHeaders`. R2 evaluates `If-None-Match: *`
atomically, so a reused URL cannot overwrite an uploaded object. The token
authenticates the server-generated key, product, expected MIME and size, and
expires after ten minutes; it is not a substitute for admin authorization.
It is signed with a key derived from the R2 secret, so rotating that secret
invalidates outstanding upload tokens.
Before issuing the PUT URL, the backend checks that PostgreSQL can select the
`confirmation_fingerprint` column defined in migration 0002. A failed check
returns 503 and does not presign. Successful checks are cached per backend
instance; failures are retried on the next request. New databases receive the
column from 0002. Databases that already applied the older 0002 must add the
nullable column and its check constraint manually before enabling uploads;
editing an applied migration file does not update their schema. This cache
assumes production migrations are not rolled back while instances are running.

The second request sends the token and optional alt text, display order, and
primary flag. The backend checks R2 metadata with HEAD, then reads at most
10 MiB from `PRODUCT_ASSETS` and fully decodes the declared PNG, JPEG, or WebP
format before inserting the image row. AVIF is temporarily suspended because
its decoder can allocate frame or tile buffers before the application checks
their dimensions. PNG, JPEG, and WebP images over 8 million pixels are rejected.
WebP uses `golang.org/x/image/webp`. The decoded dimensions are saved in
PostgreSQL. Validations are serialized per backend instance to limit
simultaneous decoder allocations. Real Worker CPU and memory behavior still
needs validation.

The 10 MiB limit is checked against the declared size before presigning and
against actual R2 metadata after PUT. The current presigned PUT does not enforce
the limit during transfer, so oversized objects can remain in R2 unregistered
and consume storage. R2 does not currently support presigned HTML-form POST
policies. If the database insert fails or confirmation never arrives, the
object also remains unregistered for safe retry. A future cleanup job should
list only keys matching `products/<product UUID>/images/<upload UUID>.<format>`
in `PRODUCT_ASSETS`, select candidates whose R2 `LastModified` is older than a
grace period well beyond the 10-minute token life (for example 24 hours), and
check each exact key against `product_images.r2_object_key` immediately before
deletion. Begin with a dry run and audit counts; recheck object metadata just
before deleting to avoid acting on a changed object. Never use a bucket-wide
lifecycle rule for these keys because registered images share the same prefix.
The cleanup design must account for database outages, in-flight confirmations,
clock differences, and any other writer with access to the bucket before it is
enabled. No automatic deletion is implemented.

The upload token expires after ten minutes. An expired token cannot create an
image, but a signed token can retrieve a previously confirmed image on retry.
The first confirmation's alt text, order, and primary choice are recorded as
an immutable fingerprint; a later confirmation with different choices receives
409, even if another image has since become primary.
Rows inserted before the fingerprint column was added have a null fingerprint.
Their repeated confirmations return 409; a recovery policy for those rows
remains to be decided.

In the Worker, the same repositories acquire a connection through the
`HYPERDRIVE` binding for each database operation. Hyperdrive performs the
connection pooling, so the Worker does not create another long-lived pool.

The `collections`, `products`, `producttypes`, `productimages`,
`productrelationships`, `productcollections`, and `productproducttypes`
PostgreSQL repositories use Bob's PostgreSQL query builder and pgx adapter. Bob
handles transactional repository operations, while the existing database
provider continues to own native pgx pooling and Cloudflare Hyperdrive
connections. Other repositories still use pgx directly and can be migrated to
Bob incrementally.

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


```sh
npx wrangler dev --config backend/wrangler.jsonc --port 8787
```
