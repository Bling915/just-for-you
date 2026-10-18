const test = require('node:test');
const assert = require('node:assert/strict');

const { createSessionStore } = require('../store/session');

function memoryStorage(initial = {}) {
  const values = new Map(Object.entries(initial));
  return {
    get: (key) => values.get(key),
    set: (key, value) => values.set(key, value),
    remove: (key) => values.delete(key),
    values,
  };
}

test('session store hydrates persisted session and user', () => {
  const storage = memoryStorage({
    session: { accessToken: 'access', refreshToken: 'refresh', userId: 7 },
    currentUser: { id: 7, displayName: '小南' },
  });
  const store = createSessionStore(storage);
  store.hydrate();
  assert.equal(store.snapshot().accessToken, 'access');
  assert.equal(store.snapshot().user.displayName, '小南');
});

test('clear removes session and user from memory and persistence', () => {
  const storage = memoryStorage();
  const store = createSessionStore(storage);
  store.setSession({ accessToken: 'access', refreshToken: 'refresh', userId: 7 });
  store.setUser({ id: 7, displayName: '小南' });
  store.clear();
  assert.deepEqual(store.snapshot(), { accessToken: '', refreshToken: '', userId: 0, user: null, epoch: 1 });
  assert.equal(storage.values.has('session'), false);
  assert.equal(storage.values.has('currentUser'), false);
});

test('a late refresh cannot restore a cleared session', () => {
  const memory = {}; const storage = { get:k=>memory[k], set:(k,v)=>memory[k]=v, remove:k=>delete memory[k] };
  const store = createSessionStore(storage);
  store.setSession({ accessToken:'old', refreshToken:'refresh', userId:1 });
  const epoch = store.snapshot().epoch;
  store.clear();
  assert.equal(store.setSession({ accessToken:'late', refreshToken:'late', userId:1 }, epoch), false);
  assert.equal(store.snapshot().accessToken, '');
});
