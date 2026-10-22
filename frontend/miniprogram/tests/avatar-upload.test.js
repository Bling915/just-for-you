const test = require('node:test');
const assert = require('node:assert/strict');
const { createAvatarReplacer } = require('../utils/avatar-upload');

function deps(overrides = {}) {
  const calls = { progress: [], updates: [] };
  return {
    calls,
    api: {
      compress: async (path) => `${path}.compressed`,
      getSize: async () => 2048,
      getMe: async () => ({ displayName: '小南', avatarRef: 'old', avatarUrl: 'old-url' }),
      requestGrant: async () => ({
        objectKey: 'users/1/avatars/new',
        uploadURL: 'https://upload',
        contentType: 'image/jpeg',
      }),
      putFile: async ({ onProgress }) => {
        onProgress(30);
        onProgress(100);
      },
      updateProfile: async (input) => {
        calls.updates.push(input);
        return { displayName: '小南', avatarRef: input.avatarRef, avatarUrl: 'new-url' };
      },
      ...overrides,
    },
  };
}

test('compresses local or chooseAvatar path, reports progress, and resolves after profile confirmation', async () => {
  const d = deps();
  const replace = createAvatarReplacer(d.api);
  const got = await replace({
    filePath: 'tmp/avatar.jpg',
    currentAvatarRef: 'old',
    userID: 1,
    onProgress: (progress) => d.calls.progress.push(progress),
  });

  assert.equal(got.avatarRef, 'users/1/avatars/new');
  assert.deepEqual(d.calls.progress, [30, 100]);
  assert.deepEqual(d.calls.updates, [{ displayName: '小南', avatarRef: 'users/1/avatars/new' }]);
});

test('grant, PUT, and profile failures never confirm a replacement', async () => {
  for (const stage of ['grant', 'put', 'profile']) {
    let updated = false;
    const overrides = stage === 'grant'
      ? { requestGrant: async () => { throw Error('grant'); } }
      : stage === 'put'
        ? { putFile: async () => { throw Error('put'); } }
        : { updateProfile: async () => { updated = true; throw Error('profile'); } };
    const d = deps(overrides);
    const replace = createAvatarReplacer(d.api);

    await assert.rejects(
      () => replace({ filePath: 'tmp/avatar.jpg', currentAvatarRef: 'old', userID: 1 }),
      { message: stage },
    );
    if (stage !== 'profile') assert.equal(d.calls.updates.length, 0);
    assert.equal(updated, stage === 'profile');
  }
});

test('rejects empty file input before requesting a grant', async () => {
  let requested = false;
  const d = deps({ requestGrant: async () => { requested = true; } });

  await assert.rejects(
    () => createAvatarReplacer(d.api)({ filePath: '', currentAvatarRef: 'old', userID: 1 }),
  );
  assert.equal(requested, false);
});
