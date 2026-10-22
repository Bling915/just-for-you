const { storage: wxStorage } = require('../utils/storage');

const EMPTY = Object.freeze({ accessToken: '', refreshToken: '', userId: 0, user: null });

function createSessionStore(storage) {
  let state = { ...EMPTY };
  let epoch = 0;
  return {
    hydrate() {
      const session = storage.get('session') || {};
      state = { ...EMPTY, ...session, user: storage.get('currentUser') || null };
      return this.snapshot();
    },
    setSession(session, expectedEpoch) {
      if (expectedEpoch !== undefined && expectedEpoch !== epoch) return false;
      state = { ...state, ...session };
      storage.set('session', {
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        userId: state.userId,
        accessExpiresAt: state.accessExpiresAt,
      });
      return expectedEpoch === undefined ? this.snapshot() : true;
    },
    setUser(user) {
      state = { ...state, user };
      storage.set('currentUser', user);
      return this.snapshot();
    },
    clear() {
      epoch += 1;
      state = { ...EMPTY };
      storage.remove('session');
      storage.remove('currentUser');
    },
    snapshot() { return { ...state, epoch }; },
  };
}

const sessionStore = createSessionStore(wxStorage);

module.exports = { createSessionStore, sessionStore };
