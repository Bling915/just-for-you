const { formatRelativeTime } = require('../../utils/date')
Component({
  properties: {
    moment: { type: Object, value: {}, observer(value){ this.setData({ displayTime: formatRelativeTime(value && value.createdAt) }) } },
    currentUserId: { type: Number, value: 0 },
  },
  data: { displayTime: '', labels: { in_class: '在上课', miss_you: '想你了', tired: '有点累' } },
  methods: {
    remove(){ this.triggerEvent('delete',{id:this.data.moment.id}) },
    preview(event){ this.triggerEvent('previewphotos',event.detail) },
  },
})
