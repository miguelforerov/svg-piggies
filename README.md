# Astro Starter Kit: Minimal

```sh
npm create astro@latest -- --template minimal
```

> 🧑‍🚀 **Seasoned astronaut?** Delete this file. Have fun!

## 🚀 Project Structure

Inside of your Astro project, you'll see the following folders and files:

```text
/
├── public/
├── src/
│   └── pages/
│       └── index.astro
└── package.json
```

Astro looks for `.astro` or `.md` files in the `src/pages/` directory. Each page is exposed as a route based on its file name.

There's nothing special about `src/components/`, but that's where we like to put any Astro/React/Vue/Svelte/Preact components.

Any static assets, like images, can be placed in the `public/` directory.

## 🧞 Commands

All commands are run from the root of the project, from a terminal:

| Command                   | Action                                           |
| :------------------------ | :----------------------------------------------- |
| `npm install`             | Installs dependencies                            |
| `npm run dev`             | Starts local dev server at `localhost:4321`      |
| `npm run build`           | Build your production site to `./dist/`          |
| `npm run preview`         | Preview your build locally, before deploying     |
| `npm run astro ...`       | Run CLI commands like `astro add`, `astro check` |
| `npm run astro -- --help` | Get help using the Astro CLI                     |

## 👀 Want to learn more?

Feel free to check [our documentation](https://docs.astro.build) or jump into our [Discord server](https://astro.build/chat).
# SVG Piggies

Digital-product storefront built with StorePlate, a Go backend on Cloudflare
Workers, Neon PostgreSQL, Cloudflare R2, and Stripe.

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
accepts connections, and applies every pending migration. The database data is
kept in the named `postgres_data` Docker volume between container restarts.

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
