Component({
  properties: {
    eyebrow: { type: String, value: 'JUST FOR US' },
    title: { type: String, value: '小事成双对' },
    currentUser: { type: Object, value: {}, observer() { this.syncInitials(); } },
    partner: { type: Object, value: {}, observer() { this.syncInitials(); } },
    days: { type: Number, value: 0 },
    soloTest: { type: Boolean, value: false },
  },
  data: { currentInitial: '我', partnerInitial: 'TA' },
  methods: {
    syncInitials() {
      const current = this.data.currentUser || {};
      const partner = this.data.partner || {};
      this.setData({
        currentInitial: (current.displayName || '我').slice(0, 1),
        partnerInitial: (partner.displayName || 'TA').slice(0, 1),
      });
    },
  },
});
