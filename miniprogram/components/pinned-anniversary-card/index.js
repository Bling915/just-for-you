Component({
  properties: {
    items: { type: Array, value: [] },
    hasMore: { type: Boolean, value: false },
  },
  methods: {
    open(e) { this.triggerEvent('open', { id: e.currentTarget.dataset.id }); },
    openAll() { this.triggerEvent('openAll'); },
  },
});