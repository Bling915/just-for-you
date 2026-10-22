const { sessionStore } = require('./store/session');
const { userService } = require('./services/users');

App({
  globalData: {
    launchedAt: Date.now(),
    reduceMotion: false,
  },
  async onLaunch() {
    try {
      const info = wx.getDeviceInfo ? wx.getDeviceInfo() : wx.getSystemInfoSync();
      this.globalData.reduceMotion = Number(info.benchmarkLevel) > 0 && Number(info.benchmarkLevel) <= 10;
    } catch (_) {}
    const session = sessionStore.hydrate();
    if (!session.accessToken) return;
    try {
      sessionStore.setUser(await userService.getMe());
    } catch (error) {
      if (error.status === 401) {
        const { appState } = require('./store/app-state');
        appState.clearPrivateData();
      }
    }
  },
});
