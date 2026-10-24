const { formatRelativeTime } = require('../../utils/date')
const { canReactToMoment, canShareToCouple } = require('../../utils/moment-actions')
Component({
  properties: {
    moment: { type: Object, value: {}, observer(){ this.refreshMoment() } },
    currentUserId: { type: Number, value: 0, observer(){ this.refreshMoment() } },
    currentUserAvatar: { type: String, value: '' },
    partnerAvatar: { type: String, value: '' },
    motionIndex: { type: Number, value: 0, observer(value){this.setData({motionDelay:Math.min(Math.max(0,Number(value)||0),8)*60,tiltClass:Number(value)%2?'tilt-right':'tilt-left'})} },
  },
  data: { displayTime: '', canReact:false,canShare:false,actionsOpen:false,motionDelay:0,tiltClass:'tilt-left' },
  lifetimes:{attached(){if(getApp().globalData.reduceMotion)this.setData({tiltClass:'reduce-motion',motionDelay:0})}},
  methods: {
    refreshMoment(){const value=this.data.moment;const index=Math.abs(Number(value&&value.id)||0)%9;this.setData({displayTime:formatRelativeTime(value&&(value.occurredAt||value.createdAt)),canReact:canReactToMoment(value,this.data.currentUserId),canShare:canShareToCouple(value,this.data.currentUserId),motionDelay:index*60,tiltClass:index%2?'tilt-right':'tilt-left'})},
    openActions(){this.setData({actionsOpen:!this.data.actionsOpen})},
    closeActions(){this.setData({actionsOpen:false})},
    avatarError(){const own=this.data.moment.authorId===this.data.currentUserId;this.setData({[own?'currentUserAvatar':'partnerAvatar']:''})},
    remove(){ this.closeActions();this.triggerEvent('delete',{id:this.data.moment.id}) },
    edit(){ this.closeActions();this.triggerEvent('edit',{moment:this.data.moment}) },
    preview(event){ this.triggerEvent('previewphotos',event.detail) },
    share(){ this.triggerEvent('share',{moment:this.data.moment}) },
    react(event){ this.triggerEvent('react',{moment:this.data.moment,...event.detail}) },
    textReaction(){ this.triggerEvent('textreaction',{moment:this.data.moment}) },
    editPerspective(){ this.triggerEvent('editperspective',{moment:this.data.moment}) },
    deletePerspective(){ this.triggerEvent('deleteperspective',{moment:this.data.moment}) },
  },
})
