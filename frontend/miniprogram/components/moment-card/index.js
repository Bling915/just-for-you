const { formatRelativeTime } = require('../../utils/date')
const { canShareToCouple } = require('../../utils/moment-actions')
Component({
  properties: {
    moment: { type: Object, value: {}, observer(value){ this.setData({ displayTime: formatRelativeTime(value && (value.occurredAt || value.createdAt)), canShare:canShareToCouple(value,this.data.currentUserId) }) } },
    currentUserId: { type: Number, value: 0 },
    currentUserAvatar: { type: String, value: '' },
    partnerAvatar: { type: String, value: '' },
  },
  data: { displayTime: '', canShare:false, labels: { in_class: '在上课', miss_you: '想你了', tired: '有点累' } },
  methods: {
    avatarError(){const own=this.data.moment.authorId===this.data.currentUserId;this.setData({[own?'currentUserAvatar':'partnerAvatar']:''})},
    remove(){ this.triggerEvent('delete',{id:this.data.moment.id}) },
    edit(){ this.triggerEvent('edit',{moment:this.data.moment}) },
    preview(event){ this.triggerEvent('previewphotos',event.detail) },
    share(){ this.triggerEvent('share',{moment:this.data.moment}) },
    react(event){ this.triggerEvent('react',{moment:this.data.moment,...event.detail}) },
    textReaction(){ this.triggerEvent('textreaction',{moment:this.data.moment}) },
  },
})
