import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';

// Wrangler accepts JSONC. Strip comments while preserving quoted URLs.
export function parseWrangler(raw) {
  const withoutComments = raw.replace(/"(?:\\.|[^"\\])*"|\/\/[^\n]*|\/\*[\s\S]*?\*\//g, (part) =>
    part.startsWith('"') ? part : '',
  );
  return JSON.parse(withoutComments.replace(/,(\s*[}\]])/g, '$1'));
}

function configured(value) {
  return typeof value === 'string' && value.trim() === value && value.length > 0 &&
    !/replace|placeholder|example|todo/i.test(value) && !/^0+$/.test(value);
}

function httpsOrigin(value, cloudflareAccess = false) {
  if (!configured(value)) return false;
  try {
    const url = new URL(value);
    return url.protocol === 'https:' && url.origin === value &&
      (!cloudflareAccess || url.hostname.endsWith('.cloudflareaccess.com'));
  } catch {
    return false;
  }
}

export function validateProductionWorker(config) {
  const production = config.env?.production;
  const vars = production?.vars;
  const required = production?.secrets?.required;
  const hyperdrive = production?.hyperdrive?.find((binding) => binding.binding === 'HYPERDRIVE');
  const assetsBucket = production?.r2_buckets?.find((binding) => binding.binding === 'PRODUCT_ASSETS');
  const filesBucket = production?.r2_buckets?.find((binding) => binding.binding === 'PRODUCT_FILES');

  if (config.workers_dev !== false || vars?.APP_ENV !== 'production' ||
    vars?.AUTH_MODE !== 'cloudflare_access' ||
    !httpsOrigin(vars?.CLOUDFLARE_ACCESS_TEAM_DOMAIN, true) ||
    !configured(vars?.CLOUDFLARE_ACCESS_AUD) ||
    vars?.DATABASE_NAME !== 'svg_piggies' ||
    !httpsOrigin(vars?.CORS_ALLOWED_ORIGIN) ||
    !configured(hyperdrive?.id) ||
    assetsBucket?.bucket_name !== 'svgpiggies-assets' ||
    vars.PRODUCT_ASSETS_BUCKET !== assetsBucket.bucket_name ||
    filesBucket?.bucket_name !== 'svgpiggies-files' ||
    vars.PRODUCT_FILES_BUCKET !== filesBucket.bucket_name ||
    !Array.isArray(required) ||
    !['ADMIN_EMAILS', 'STRIPE_SECRET_KEY', 'STRIPE_WEBHOOK_SECRET', 'R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_ENDPOINT'].every(
      (name) => required.includes(name))) {
    throw new Error('Unsafe or incomplete production Worker configuration; deployment stopped.');
  }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const raw = readFileSync(new URL('../backend/wrangler.jsonc', import.meta.url), 'utf8');
  validateProductionWorker(parseWrangler(raw));
}
