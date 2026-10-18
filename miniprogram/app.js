const { sessionStore } = require('./store/session');
const { userService } = require('./services/users');

App({
  globalData: {
    launchedAt: Date.now(),
  },
  async onLaunch() {
    const session = sessionStore.hydrate();
    if (!session.accessToken) return;
    try {
      sessionStore.setUser(await userService.getMe());
    } catch (error) {
      if (error.status === 401) sessionStore.clear();
    }
  },
});
