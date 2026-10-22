const { userService } = require('../../services/users');
const { sessionStore } = require('../../store/session');
const { messageForError } = require('../../utils/error-copy');
const { countRunes, truncateRunes } = require('../../utils/runes');

Page({
  data: { displayName: '', initial: '♡', user: {}, submitting: false, error: '', editing: false },
  onLoad(options) { const user=sessionStore.snapshot().user||{}; const editing=options&&options.edit==='1'; const displayName=editing?(user.displayName||''):''; this.setData({editing,user,displayName,initial:displayName.slice(0,1)||'♡'}); },
  avatarChanged(event) { this.setData({user:event.detail.user}); },
  inputName(event) { const displayName=truncateRunes(event.detail.value,30);this.setData({displayName,initial:displayName.slice(0,1)||'♡',error:''}); },
  async submit() {
    if (this.data.submitting || countRunes(this.data.displayName.trim()) < 1) {
      this.setData({ error: '请填写 1 至 30 个字符的昵称' });
      return;
    }
    this.setData({ submitting: true, error: '' });
    try {
      const user = await userService.updateProfile({ displayName: this.data.displayName.trim(), avatarRef: this.data.user.avatarRef || '' });
      sessionStore.setUser(user);
      if (this.data.editing) wx.navigateBack(); else wx.redirectTo({ url: '/pages/bind/index' });
    } catch (error) {
      this.setData({ error: error.code === 'INVALID_PROFILE' ? '请填写 1 至 30 个字符的昵称' : messageForError(error.code) });
    } finally { this.setData({ submitting: false }); }
  },
});
