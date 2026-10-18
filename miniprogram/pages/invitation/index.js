const { relationshipService } = require('../../services/relationships');
const { messageForError } = require('../../utils/error-copy');

Page({
  data: { code: '', expiresAt: '', created: false, invitation: null, loading: true, submitting: false, error: '' },
  async onLoad(options) {
    const code = String(options.code || '');
    const created = options.created === '1';
    this.setData({ code, created, expiresAt: options.expiresAt || '', loading: !created });
    if (created) return;
    try { this.setData({ invitation: await relationshipService.previewInvitation(code) }); }
    catch (error) { this.setData({ error: messageForError(error.code) }); }
    finally { this.setData({ loading: false }); }
  },
  copy() { wx.setClipboardData({ data: this.data.code }); },
  async bind() {
    if (this.data.submitting) return;
    this.setData({ submitting: true, error: '' });
    try {
      await relationshipService.bind(this.data.code);
      wx.showToast({ title: '绑定成功', icon: 'success' });
      setTimeout(() => wx.reLaunch({ url: '/pages/bind/index' }), 500);
    } catch (error) { this.setData({ error: messageForError(error.code) }); }
    finally { this.setData({ submitting: false }); }
  },
});
