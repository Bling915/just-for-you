Component({
  properties: { value: { type: String, value: 'shared' }, hasPartner: { type: Boolean, value: true } },
  methods: {
    choose(event) {
      const value = event.currentTarget.dataset.value
      this.triggerEvent('change', { value })
    },
  },
})
