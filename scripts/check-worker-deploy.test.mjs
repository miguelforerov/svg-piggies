import assert from 'node:assert/strict';
import test from 'node:test';
import { parseWrangler, validateProductionWorker } from './check-worker-deploy.mjs';

function safeConfig() {
  return {
    workers_dev: false,
    env: { production: {
      vars: { APP_ENV: 'production', AUTH_MODE: 'cloudflare_access',
        CLOUDFLARE_ACCESS_TEAM_DOMAIN: 'https://team.cloudflareaccess.com',
        CLOUDFLARE_ACCESS_AUD: 'application-audience', DATABASE_NAME: 'svg_piggies',
        CORS_ALLOWED_ORIGIN: 'https://store.svgpiggies.com', PRODUCT_ASSETS_BUCKET: 'svgpiggies-assets',
        PRODUCT_FILES_BUCKET: 'svgpiggies-files' },
      secrets: { required: ['ADMIN_EMAILS', 'STRIPE_SECRET_KEY', 'STRIPE_WEBHOOK_SECRET',
        'R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_ENDPOINT'] },
      hyperdrive: [{ binding: 'HYPERDRIVE', id: '123456789abcdef0123456789abcdef0' }],
      r2_buckets: [{ binding: 'PRODUCT_ASSETS', bucket_name: 'svgpiggies-assets' },
        { binding: 'PRODUCT_FILES', bucket_name: 'svgpiggies-files' }],
    } },
  };
}

test('parses JSONC without changing URLs', () => {
  const config = parseWrangler('{ // comment\n "url": "https://example.invalid", }');
  assert.equal(config.url, 'https://example.invalid');
});

test('accepts a complete protected production configuration', () => {
  assert.doesNotThrow(() => validateProductionWorker(safeConfig()));
});

test('blocks unsafe Worker deploy configurations', () => {
  for (const edit of [
    (c) => { c.workers_dev = true; },
    (c) => { c.env.production.vars.APP_ENV = 'development'; },
    (c) => { c.env.production.vars.AUTH_MODE = 'development'; },
    (c) => { c.env.production.vars.CLOUDFLARE_ACCESS_AUD = 'replace-with-aud'; },
    (c) => { c.env.production.vars.CLOUDFLARE_ACCESS_TEAM_DOMAIN = 'https://other.test'; },
    (c) => { c.env.production.vars.CORS_ALLOWED_ORIGIN = 'https://replace-with-store.test'; },
    (c) => { c.env.production.secrets.required = []; },
    (c) => { c.env.production.secrets.required = ['ADMIN_EMAILS']; },
    (c) => { c.env.production.hyperdrive[0].id = '00000000000000000000000000000000'; },
    (c) => { c.env.production.hyperdrive[0].id = 'replace-with-hyperdrive-id'; },
    (c) => { c.env.production.vars.PRODUCT_ASSETS_BUCKET = 'svgpiggies-assets-dev'; },
    (c) => { c.env.production.r2_buckets[1].bucket_name = 'svgpiggies-files-dev'; },
  ]) {
    const config = safeConfig();
    edit(config);
    assert.throws(() => validateProductionWorker(config));
  }
});
