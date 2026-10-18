const { authService } = require('../../services/auth');
const { userService } = require('../../services/users');
const { sessionStore } = require('../../store/session');
const { messageForError } = require('../../utils/error-copy');
const { appState } = require('../../store/app-state');

Page({
  data: { submitting: false, error: '' },
  async login() {
    if (this.data.submitting) return;
    this.setData({ submitting: true, error: '' });
    try {
      const login = await wx.login();
      const session = await authService.loginWithWeChat(login.code);
      appState.clearPrivateData();
      sessionStore.setSession(session);
      const user = await userService.getMe();
      sessionStore.setUser(user);
      wx.redirectTo({ url: user.displayName ? '/pages/bind/index' : '/pages/profile-setup/index' });
    } catch (error) {
      this.setData({ error: messageForError(error.code) });
    } finally {
      this.setData({ submitting: false });
    }
  },
});
