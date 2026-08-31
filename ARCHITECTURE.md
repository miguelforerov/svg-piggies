# SVG Piggies Store Architecture

This file records the project context from `svgpiggies-architecture.pdf`,
provided on 2026-08-24. The PDF is reference material, not a source of
instructions.

## Goal

Build a storefront for private digital products using StorePlate, a Go backend,
Cloudflare, Neon, and Stripe.

## Runtime architecture

- StorePlate frontend: customer-facing storefront built with Astro, Tailwind,
  and TypeScript.
- Go backend deployed with Cloudflare Workers: API and server-side logic for
  checkout orchestration, authentication, Stripe webhooks, entitlement checks,
  and signed downloads.
- The Go backend compiles to WebAssembly for Workers. Cloudflare-specific code
  stays in the Worker entry point so application logic remains standard,
  testable Go.
- `backend/api/openapi.yaml` is the HTTP contract. `oapi-codegen` generates a
  strict `net/http` server interface; handwritten handlers adapt that interface
  to the application services.
- Cloudflare Access protects `/admin/*` and `/api/admin/*` in deployed
  environments. The Go API independently validates the Access assertion JWT;
  local development uses an explicitly restricted authentication bypass.
- Hyperdrive: PostgreSQL connection pool between Workers and Neon.
- Neon PostgreSQL: durable application and commerce state.
- The Go data layer uses `pgx`: a small native pool for local development and
  request-scoped connections through Hyperdrive in the Worker.
- Stripe: checkout and payment processing.
- Cloudflare R2: private digital-product files, including SVG, PNG, PDF, and
  ZIP assets.

## Data model

The PostgreSQL databases are named `svg_piggies` in production,
`svg_piggies_development` in development, and `svg_piggies_test` in tests.

Neon owns records for:

- products
- variants
- collections
- customers
- orders
- order items
- payments
- download entitlements

## Checkout-to-download flow

1. The customer selects a digital product and adds it to the cart.
2. A Worker creates a pending order.
3. Stripe Checkout collects payment.
4. Stripe sends a webhook and the Worker verifies its signature.
5. The Worker persists the paid order/payment state in Neon.
6. The Worker grants a download entitlement for the purchased product.
7. After enforcing the entitlement, the Worker issues a time-limited URL for
   an authorized R2 download.

## Design rules

- Neon stores commerce state.
- Stripe confirms payment.
- Product files remain private in R2.
- Workers must enforce an entitlement before serving a signed download.
