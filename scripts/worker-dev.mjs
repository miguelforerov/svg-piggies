import { spawn } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseEnv } from 'node:util';
import { fileURLToPath } from 'node:url';
import { parseWrangler } from './check-worker-deploy.mjs';

const root = fileURLToPath(new URL('../', import.meta.url));
const hyperdriveVariable = 'CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_HYPERDRIVE';
const requiredFromEnv = [
  'APP_ENV', 'AUTH_MODE', 'DATABASE_NAME', 'PRODUCT_ASSETS_BUCKET', 'PRODUCT_FILES_BUCKET',
  'R2_REGION', 'R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_ENDPOINT',
  hyperdriveVariable, 'PRODUCT_FILE_KEYS_JSON', 'PRODUCT_FILE_ACTIVE_KEY_VERSION',
];
const requiredWorkerSecrets = [
  'R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_ENDPOINT',
  'PRODUCT_FILE_KEYS_JSON', 'PRODUCT_FILE_ACTIVE_KEY_VERSION',
];

export const wranglerDevArgs = [
  fileURLToPath(new URL('../node_modules/wrangler/bin/wrangler.js', import.meta.url)),
  'dev', '--config', 'backend/wrangler.jsonc', '--local', '--var', 'WORKER_LOCAL_DEV:true',
];

function invalid(name, reason) {
  throw new Error(`${name}: ${reason}`);
}

function requireValue(values, name) {
  if (typeof values[name] !== 'string' || values[name].trim() === '') {
    invalid(name, 'required for local Worker development');
  }
}

function requireExact(values, name, expected) {
  if (values[name] !== expected) invalid(name, 'does not match the development configuration');
}

function validateLocalDatabase(raw) {
  let url;
  try {
    url = new URL(raw);
  } catch {
    invalid(hyperdriveVariable, 'must be a local PostgreSQL URL');
  }
  let databaseName;
  try {
    databaseName = decodeURIComponent(url.pathname);
  } catch {
    invalid(hyperdriveVariable, 'has an invalid database name');
  }
  if (!['postgres:', 'postgresql:'].includes(url.protocol) ||
      !['localhost', '127.0.0.1', '[::1]'].includes(url.hostname) ||
      url.hash || [...url.searchParams.keys()].some((name) => name !== 'sslmode') ||
      url.searchParams.getAll('sslmode').length > 1) {
    invalid(hyperdriveVariable, 'must use a local PostgreSQL host');
  }
  if (databaseName !== '/svg_piggies_development') {
    invalid(hyperdriveVariable, 'must select svg_piggies_development');
  }
}

function validateR2Endpoint(raw) {
  let url;
  try {
    url = new URL(raw);
  } catch {
    invalid('R2_ENDPOINT', 'must be an HTTPS origin');
  }
  if (url.protocol !== 'https:' || !url.hostname || url.username || url.password ||
      (url.pathname !== '/' && url.pathname !== '') || url.search || url.hash) {
    invalid('R2_ENDPOINT', 'must be an HTTPS origin without credentials or a path');
  }
}

function validateProductFileKeys(env) {
  const version = env.PRODUCT_FILE_ACTIVE_KEY_VERSION;
  if (!/^[1-9][0-9]*$/.test(version) || Number(version) > 2147483647) {
    invalid('PRODUCT_FILE_ACTIVE_KEY_VERSION', 'must be a positive key version');
  }
  let keys;
  try {
    keys = JSON.parse(env.PRODUCT_FILE_KEYS_JSON);
  } catch {
    invalid('PRODUCT_FILE_KEYS_JSON', 'must be valid JSON');
  }
  if (keys === null || Array.isArray(keys) || typeof keys !== 'object' ||
      !Object.hasOwn(keys, version)) {
    invalid('PRODUCT_FILE_KEYS_JSON', 'must contain the active key version');
  }
  for (const [keyVersion, pair] of Object.entries(keys)) {
    if (!/^[1-9][0-9]*$/.test(keyVersion) || Number(keyVersion) > 2147483647 ||
        pair === null || Array.isArray(pair) || typeof pair !== 'object') {
      invalid('PRODUCT_FILE_KEYS_JSON', 'contains an invalid key version');
    }
    for (const name of ['signing', 'fingerprint']) {
      const encoded = pair[name];
      if (typeof encoded !== 'string' || !/^[A-Za-z0-9+/]+={0,2}$/.test(encoded) ||
          Buffer.from(encoded, 'base64').length !== 32 ||
          Buffer.from(encoded, 'base64').toString('base64') !== encoded) {
        invalid('PRODUCT_FILE_KEYS_JSON', 'requires base64 encoded 32-byte keys');
      }
    }
    if (pair.signing === pair.fingerprint) {
      invalid('PRODUCT_FILE_KEYS_JSON', 'requires distinct signing and fingerprint keys');
    }
  }
}

export function validateLocalWorkerEnvironment(env, envFile, wranglerConfig) {
  if (env.CLOUDFLARE_ENV) invalid('CLOUDFLARE_ENV', 'must be unset for the local Worker');
  if (env.WRANGLER_HYPERDRIVE_LOCAL_CONNECTION_STRING_HYPERDRIVE) {
    invalid('WRANGLER_HYPERDRIVE_LOCAL_CONNECTION_STRING_HYPERDRIVE', 'deprecated override must be unset');
  }
  if (env.CLOUDFLARE_LOAD_DEV_VARS_FROM_DOT_ENV?.toLowerCase() === 'false') {
    invalid('CLOUDFLARE_LOAD_DEV_VARS_FROM_DOT_ENV', 'must not disable local secrets');
  }
  for (const name of requiredFromEnv) {
    requireValue(envFile, name);
    requireValue(env, name);
    if (env[name] !== envFile[name]) invalid(name, 'inherited environment overrides .env');
  }

  const expected = {
    APP_ENV: 'development', AUTH_MODE: 'development', DATABASE_NAME: 'svg_piggies_development',
    PRODUCT_ASSETS_BUCKET: 'svgpiggies-assets-development',
    PRODUCT_FILES_BUCKET: 'svgpiggies-files-development', R2_REGION: 'auto',
  };
  for (const [name, value] of Object.entries(expected)) {
    requireExact(env, name, value);
    requireExact(wranglerConfig.vars ?? {}, name, value);
  }
  for (const [binding, bucket] of [
    ['PRODUCT_ASSETS', expected.PRODUCT_ASSETS_BUCKET],
    ['PRODUCT_FILES', expected.PRODUCT_FILES_BUCKET],
  ]) {
    if (wranglerConfig.r2_buckets?.find((entry) => entry.binding === binding)?.bucket_name !== bucket) {
      invalid(binding, 'must bind the development bucket');
    }
  }
  if (!wranglerConfig.hyperdrive?.some((entry) => entry.binding === 'HYPERDRIVE')) {
    invalid('HYPERDRIVE', 'binding is required');
  }
  for (const name of requiredWorkerSecrets) {
    if (!wranglerConfig.secrets?.required?.includes(name)) {
      invalid(name, 'must be declared in local Worker secrets.required');
    }
  }
  for (const name of ['CORS_ALLOWED_ORIGIN', 'DOWNLOAD_URL_TTL_SECONDS']) {
    if (env[name] && env[name] !== wranglerConfig.vars?.[name]) {
      invalid(name, 'inherited environment overrides the Worker configuration');
    }
  }
  if (env.WORKER_LOCAL_DEV && env.WORKER_LOCAL_DEV !== 'true') {
    invalid('WORKER_LOCAL_DEV', 'conflicts with the local Worker command');
  }
  validateLocalDatabase(env[hyperdriveVariable]);
  validateR2Endpoint(env.R2_ENDPOINT);
  validateProductFileKeys(env);
}

function readLocalConfiguration() {
  let source;
  try {
    source = readFileSync(new URL('../.env', import.meta.url), 'utf8');
  } catch {
    invalid('.env', 'missing or unreadable root development file');
  }
  const names = new Set();
  for (const line of source.split(/\r?\n/)) {
    const name = line.match(/^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=/)?.[1];
    if (name && names.has(name)) invalid(name, 'duplicate definition in .env');
    if (name) names.add(name);
  }
  let envFile;
  try {
    envFile = parseEnv(source);
  } catch {
    invalid('.env', 'invalid root development file');
  }
  let wranglerConfig;
  try {
    wranglerConfig = parseWrangler(readFileSync(new URL('../backend/wrangler.jsonc', import.meta.url), 'utf8'));
  } catch {
    invalid('backend/wrangler.jsonc', 'missing or invalid Worker configuration');
  }
  return { envFile, wranglerConfig };
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const { envFile, wranglerConfig } = readLocalConfiguration();
    validateLocalWorkerEnvironment(process.env, envFile, wranglerConfig);
    const child = spawn(process.execPath, wranglerDevArgs, { cwd: root, stdio: 'inherit' });
    for (const signal of ['SIGINT', 'SIGTERM']) process.on(signal, () => child.kill(signal));
    child.on('error', () => {
      console.error('wrangler: failed to start the local Worker');
      process.exitCode = 1;
    });
    child.on('exit', (code) => { process.exitCode = code ?? 1; });
  } catch (error) {
    console.error(`worker:dev preflight: ${error.message}`);
    process.exitCode = 1;
  }
}
