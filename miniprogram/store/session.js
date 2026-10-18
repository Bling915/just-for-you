const { storage: wxStorage } = require('../utils/storage');

const EMPTY = Object.freeze({ accessToken: '', refreshToken: '', userId: 0, user: null });

function createSessionStore(storage) {
  let state = { ...EMPTY };
  return {
    hydrate() {
      const session = storage.get('session') || {};
      state = { ...EMPTY, ...session, user: storage.get('currentUser') || null };
      return this.snapshot();
    },
    setSession(session) {
      state = { ...state, ...session };
      storage.set('session', {
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        userId: state.userId,
        accessExpiresAt: state.accessExpiresAt,
      });
      return this.snapshot();
    },
    setUser(user) {
      state = { ...state, user };
      storage.set('currentUser', user);
      return this.snapshot();
    },
    clear() {
      state = { ...EMPTY };
      storage.remove('session');
      storage.remove('currentUser');
    },
    snapshot() { return { ...state }; },
  };
}

const sessionStore = createSessionStore(wxStorage);

module.exports = { createSessionStore, sessionStore };
