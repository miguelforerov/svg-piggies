import assert from 'node:assert/strict';
import test from 'node:test';
import { validateLocalWorkerEnvironment } from './worker-dev.mjs';

const hyperdrive = 'CLOUDFLARE_HYPERDRIVE_LOCAL_CONNECTION_STRING_HYPERDRIVE';

function safeFixture() {
  const fileVars = {
    APP_ENV: 'development', AUTH_MODE: 'development', DATABASE_NAME: 'svg_piggies_development',
    PRODUCT_ASSETS_BUCKET: 'svgpiggies-assets-development',
    PRODUCT_FILES_BUCKET: 'svgpiggies-files-development', R2_REGION: 'auto',
    R2_ACCESS_KEY_ID: 'fake-local-id', R2_SECRET_ACCESS_KEY: 'fake-local-secret',
    R2_ENDPOINT: 'https://fake-account.r2.cloudflarestorage.com',
    [hyperdrive]: 'postgresql://fake:fake@127.0.0.1:5432/svg_piggies_development?sslmode=disable',
    PRODUCT_FILE_ACTIVE_KEY_VERSION: '1',
    PRODUCT_FILE_KEYS_JSON: JSON.stringify({
      1: { signing: Buffer.alloc(32, 1).toString('base64'),
        fingerprint: Buffer.alloc(32, 2).toString('base64') },
    }),
  };
  const config = {
    vars: { APP_ENV: fileVars.APP_ENV, AUTH_MODE: fileVars.AUTH_MODE,
      DATABASE_NAME: fileVars.DATABASE_NAME, PRODUCT_ASSETS_BUCKET: fileVars.PRODUCT_ASSETS_BUCKET,
      PRODUCT_FILES_BUCKET: fileVars.PRODUCT_FILES_BUCKET, R2_REGION: fileVars.R2_REGION },
    r2_buckets: [
      { binding: 'PRODUCT_ASSETS', bucket_name: fileVars.PRODUCT_ASSETS_BUCKET },
      { binding: 'PRODUCT_FILES', bucket_name: fileVars.PRODUCT_FILES_BUCKET },
    ],
    hyperdrive: [{ binding: 'HYPERDRIVE', id: 'local-placeholder' }],
    secrets: { required: ['R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_ENDPOINT',
      'PRODUCT_FILE_KEYS_JSON', 'PRODUCT_FILE_ACTIVE_KEY_VERSION'] },
  };
  return { env: { ...fileVars }, fileVars, config };
}

test('accepts the exact local configuration', () => {
  const { env, fileVars, config } = safeFixture();
  assert.doesNotThrow(() => validateLocalWorkerEnvironment(env, fileVars, config));
});

test('rejects missing required variables and R2 credentials', () => {
  for (const name of ['APP_ENV', 'R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_ENDPOINT', hyperdrive]) {
    const { env, fileVars, config } = safeFixture();
    delete env[name];
    assert.throws(() => validateLocalWorkerEnvironment(env, fileVars, config),
      (error) => error.message.startsWith(`${name}:`));
  }
});

test('rejects an incorrect environment or either bucket', () => {
  for (const [name, value] of [
    ['APP_ENV', 'production'],
    ['PRODUCT_ASSETS_BUCKET', 'svgpiggies-assets'],
    ['PRODUCT_FILES_BUCKET', 'svgpiggies-files'],
    ['PRODUCT_FILES_BUCKET', 'another-bucket'],
  ]) {
    const { env, fileVars, config } = safeFixture();
    env[name] = fileVars[name] = value;
    assert.throws(() => validateLocalWorkerEnvironment(env, fileVars, config),
      (error) => error.message.startsWith(`${name}:`));
  }
});

test('rejects a non-local PostgreSQL host or a different database', () => {
  for (const value of [
    'postgresql://fake:fake@database.example/svg_piggies_development',
    'postgresql://fake:fake@127.0.0.1/svg_piggies',
    'postgresql://fake:fake@127.0.0.1/svg_piggies_development?host=database.example',
    'postgresql://fake:fake@127.0.0.1/svg_piggies_development?hostaddr=192.0.2.1',
  ]) {
    const { env, fileVars, config } = safeFixture();
    env[hyperdrive] = fileVars[hyperdrive] = value;
    assert.throws(() => validateLocalWorkerEnvironment(env, fileVars, config),
      (error) => error.message.startsWith(`${hyperdrive}:`));
  }
});

test('rejects missing, malformed, or reused upload authorization keys', () => {
  for (const edit of [
    (f) => { delete f.env.PRODUCT_FILE_KEYS_JSON; },
    (f) => { delete f.env.PRODUCT_FILE_ACTIVE_KEY_VERSION; },
    (f) => { f.env.PRODUCT_FILE_KEYS_JSON = f.fileVars.PRODUCT_FILE_KEYS_JSON = '{'; },
    (f) => { f.env.PRODUCT_FILE_ACTIVE_KEY_VERSION = f.fileVars.PRODUCT_FILE_ACTIVE_KEY_VERSION = '2'; },
    (f) => { f.env.PRODUCT_FILE_KEYS_JSON = f.fileVars.PRODUCT_FILE_KEYS_JSON =
      JSON.stringify({ 1: { signing: 'bad', fingerprint: 'bad' } }); },
    (f) => { const pair = JSON.parse(f.env.PRODUCT_FILE_KEYS_JSON);
      pair[1].fingerprint = pair[1].signing;
      f.env.PRODUCT_FILE_KEYS_JSON = f.fileVars.PRODUCT_FILE_KEYS_JSON = JSON.stringify(pair); },
  ]) {
    const fixture = safeFixture();
    edit(fixture);
    assert.throws(() => validateLocalWorkerEnvironment(fixture.env, fixture.fileVars, fixture.config),
      (error) => /^PRODUCT_FILE_(KEYS_JSON|ACTIVE_KEY_VERSION):/.test(error.message));
  }
});

test('rejects inherited overrides without exposing values', () => {
  const { env, fileVars, config } = safeFixture();
  env.R2_SECRET_ACCESS_KEY = 'inherited-other-secret';
  assert.throws(() => validateLocalWorkerEnvironment(env, fileVars, config),
    (error) => error.message.startsWith('R2_SECRET_ACCESS_KEY:') &&
      !error.message.includes('inherited-other-secret'));
});

test('rejects an inherited Cloudflare environment or disabled local secrets', () => {
  for (const [name, value] of [
    ['CLOUDFLARE_ENV', 'production'],
    ['CLOUDFLARE_LOAD_DEV_VARS_FROM_DOT_ENV', 'false'],
  ]) {
    const { env, fileVars, config } = safeFixture();
    env[name] = value;
    assert.throws(() => validateLocalWorkerEnvironment(env, fileVars, config),
      (error) => error.message.startsWith(`${name}:`));
  }
});

test('rejects a production binding or a missing Worker secret declaration', () => {
  const fixture = safeFixture();
  fixture.config.r2_buckets[0].bucket_name = 'svgpiggies-assets';
  assert.throws(() => validateLocalWorkerEnvironment(fixture.env, fixture.fileVars, fixture.config),
    (error) => error.message.startsWith('PRODUCT_ASSETS:'));
  const other = safeFixture();
  other.config.secrets.required.pop();
  assert.throws(() => validateLocalWorkerEnvironment(other.env, other.fileVars, other.config),
    (error) => error.message.startsWith('PRODUCT_FILE_ACTIVE_KEY_VERSION:'));
});
