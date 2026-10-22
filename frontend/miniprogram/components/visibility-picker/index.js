Component({
  properties: { value: { type: String, value: 'private' }, hasPartner: { type: Boolean, value: true } },
  methods: {
    choose(event) {
      const value = event.currentTarget.dataset.value
      if (value === 'partner' && !this.data.hasPartner) {
        wx.showToast({ title: '绑定伴侣后才能分享给 TA', icon: 'none' })
        return
      }
      this.triggerEvent('change', { value })
    },
  },
})
