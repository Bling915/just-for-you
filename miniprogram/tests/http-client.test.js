const test = require('node:test');
const assert = require('node:assert/strict');

const { createHttpClient } = require('../services/http-client');

function createSession(initial) {
  let value = { ...initial };
  return {
    snapshot: () => ({ ...value }),
    setSession: (next) => { value = { ...next }; },
    clear: () => { value = {}; },
  };
}

test('client unwraps data and sends the access token', async () => {
  const calls = [];
  const client = createHttpClient({
    baseURL: 'https://api.test',
    storage: createSession({ accessToken: 'access', refreshToken: 'refresh' }),
    onSessionExpired() {},
    request: async (options) => {
      calls.push(options);
      return { statusCode: 200, data: { data: { ok: true } } };
    },
  });
  assert.deepEqual(await client.get('/v1/home'), { ok: true });
  assert.equal(calls[0].header.Authorization, 'Bearer access');
});

test('client exposes stable business error fields', async () => {
  const client = createHttpClient({
    baseURL: 'https://api.test', storage: createSession({}), onSessionExpired() {},
    request: async () => ({ statusCode: 400, data: { error: { code: 'EMPTY_MOMENT', message: '请记录一些内容' } } }),
  });
  await assert.rejects(client.post('/v1/moments', {}), (error) => {
    assert.deepEqual(error, { status: 400, code: 'EMPTY_MOMENT', message: '请记录一些内容' });
    return true;
  });
});

test('three concurrent 401 responses share one refresh and replay all requests', async () => {
  const session = createSession({ accessToken: 'old', refreshToken: 'refresh' });
  let refreshCalls = 0;
  let protectedCalls = 0;
  const request = async (options) => {
    if (options.url.endsWith('/v1/auth/refresh')) {
      refreshCalls += 1;
      await new Promise((resolve) => setTimeout(resolve, 10));
      return { statusCode: 200, data: { data: { accessToken: 'new', refreshToken: 'next', userId: 1 } } };
    }
    protectedCalls += 1;
    if (options.header.Authorization === 'Bearer old') return { statusCode: 401, data: { error: { code: 'UNAUTHORIZED' } } };
    return { statusCode: 200, data: { data: options.url } };
  };
  const client = createHttpClient({ baseURL: 'https://api.test', request, storage: session, onSessionExpired() {} });
  const values = await Promise.all(['/a', '/b', '/c'].map((path) => client.get(path)));
  assert.equal(refreshCalls, 1);
  assert.equal(protectedCalls, 6);
  assert.deepEqual(values, ['https://api.test/a', 'https://api.test/b', 'https://api.test/c']);
});

test('failed shared refresh expires the session once', async () => {
  const session = createSession({ accessToken: 'old', refreshToken: 'refresh' });
  let expiredCalls = 0;
  let refreshCalls = 0;
  const request = async (options) => {
    if (options.url.endsWith('/v1/auth/refresh')) {
      refreshCalls += 1;
      await new Promise((resolve) => setTimeout(resolve, 10));
      return { statusCode: 401, data: { error: { code: 'INVALID_REFRESH' } } };
    }
    return { statusCode: 401, data: { error: { code: 'UNAUTHORIZED' } } };
  };
  const client = createHttpClient({ baseURL: 'https://api.test', request, storage: session, onSessionExpired: () => { expiredCalls += 1; } });
  const results = await Promise.allSettled(['/a', '/b', '/c'].map((path) => client.get(path)));
  assert.ok(results.every((result) => result.status === 'rejected'));
  assert.equal(refreshCalls, 1);
  assert.equal(expiredCalls, 1);
});
