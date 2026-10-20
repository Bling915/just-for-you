const { relationshipService } = require('../../services/relationships');
const { normalizeInvitationCode, messageForError } = require('../../utils/error-copy');
const { APP_CONFIG } = require('../../config');

Page({
  data: { code: '', creating: false, checking: false, soloTesting: false, soloTestEnabled: APP_CONFIG.soloTestEnabled, error: '', relationship: null },
  async onLoad() {
    try { this.setData({ relationship: await relationshipService.getCurrent() }); return wx.switchTab({url:'/pages/home/index'}); } catch (error) {
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
  async enableSoloTest() {
    if (this.data.soloTesting) return;
    const confirmation = await wx.showModal({ title: '开启单人体验？', content: '系统会为你创建一位虚拟搭子，仅用于测试当前版本。', confirmText: '开始体验' });
    if (!confirmation.confirm) return;
    this.setData({ soloTesting: true, error: '' });
    try {
      await relationshipService.enableSoloTest();
      wx.switchTab({ url: '/pages/home/index' });
    } catch (error) {
      this.setData({ error: messageForError(error.code) });
    } finally {
      this.setData({ soloTesting: false });
    }
  },
});
