Component({
  properties: { photos: { type: Array, value: [] } },
  data: { failed: {} },
  methods: {
    preview(event) { const index=Number(event.currentTarget.dataset.index); const urls=this.data.photos.map((p)=>p.url).filter(Boolean); if(urls[index]) this.triggerEvent('preview',{current:urls[index],urls}) },
    fail(event) { this.setData({ [`failed.${event.currentTarget.dataset.index}`]: true }) },
  },
})
