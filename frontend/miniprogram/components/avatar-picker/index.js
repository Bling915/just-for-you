const { replaceAvatar } = require('../../utils/avatar-upload');
const { sessionStore } = require('../../store/session');

Component({
  properties: {
    user: { type: Object, value: {}, observer(user) { this.setData({ initial: ((user && user.displayName) || '♡').slice(0, 1) }); } },
    size: { type: String, value: 'large' },
  },
  data: { uploading: false, progress: 0, initial: '♡' },
  methods: {
    chooseAvatar(event) {
      const filePath = event.detail && event.detail.avatarUrl;
      if (filePath) this.upload(filePath);
    },
    chooseFromAlbum() {
      if (this.data.uploading) return;
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: (result) => this.upload(result.tempFiles[0].tempFilePath),
      });
    },
    async upload(filePath) {
      if (this.data.uploading) return;
      this.setData({ uploading: true, progress: 0 });
      try {
        const user = await replaceAvatar({
          filePath,
          onProgress: (progress) => this.setData({ progress }),
        });
        sessionStore.setUser(user);
        this.triggerEvent('change', { user });
        wx.showToast({ title: '头像换好啦', icon: 'success' });
      } catch (error) {
        wx.showToast({
          title: error.code === 'INVALID_AVATAR' ? '请选择 5MB 以内的图片' : '头像上传失败，请重试',
          icon: 'none',
        });
      } finally {
        this.setData({ uploading: false, progress: 0 });
      }
    },
  },
});
