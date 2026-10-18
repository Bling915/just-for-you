const { relationshipService } = require('../../services/relationships');
const { normalizeInvitationCode, messageForError } = require('../../utils/error-copy');

Page({
  data: { code: '', creating: false, checking: false, error: '', relationship: null },
  async onLoad() {
    try { this.setData({ relationship: await relationshipService.getCurrent() }); } catch (error) {
      if (error.code !== 'NO_ACTIVE_COUPLE') this.setData({ error: messageForError(error.code) });
    }
  },
  inputCode(event) { this.setData({ code: normalizeInvitationCode(event.detail.value), error: '' }); },
  async createInvitation() {
    if (this.data.creating) return;
    this.setData({ creating: true, error: '' });
    try {
      const invitation = await relationshipService.createInvitation();
      wx.navigateTo({ url: `/pages/invitation/index?created=1&code=${encodeURIComponent(invitation.code)}&expiresAt=${encodeURIComponent(invitation.expiresAt)}` });
    } catch (error) { this.setData({ error: messageForError(error.code) }); }
    finally { this.setData({ creating: false }); }
  },
  preview() {
    if (this.data.checking || this.data.code.length !== 6) { this.setData({ error: '请输入完整的六位邀请码' }); return; }
    wx.navigateTo({ url: `/pages/invitation/index?code=${encodeURIComponent(this.data.code)}` });
  },
});
