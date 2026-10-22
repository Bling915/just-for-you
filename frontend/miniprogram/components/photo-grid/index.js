Component({
  properties: {
    photos: { type: Array, value: [] },
  },
  data: { failed: {} },
  observers: {
    // photos 变化说明拿到了新的签名 URL，必须清空失败标记，
    // 否则组件复用旧索引导致图片永远显示占位文案。
    'photos.**': function () {
      if (Object.keys(this.data.failed).length) this.setData({ failed: {} });
    },
  },
  methods: {
    preview(event) {
      const index = Number(event.currentTarget.dataset.index);
      const urls = this.data.photos.map((p) => p.url).filter(Boolean);
      if (urls[index]) this.triggerEvent('preview', { current: urls[index], urls });
    },
    fail(event) {
      this.setData({ [`failed.${event.currentTarget.dataset.index}`]: true });
    },
  },
});