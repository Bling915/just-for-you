const { reactionView } = require('../../utils/moment-actions')
Component({
  properties: {
    reactions: { type: Array, value: [], observer(){ this.refresh() } },
    currentUserId: { type: Number, value: 0, observer(){ this.refresh() } },
  },
  data: { items: [], presets: [{ type: 'seen', icon: '👀' }, { type: 'hug', icon: '🫂' }, { type: 'together', icon: '✨' }, { type: 'collect', icon: '♡' }] },
  lifetimes: { attached(){ this.refresh() } },
  methods: {
    refresh(){ this.setData({ items: reactionView(this.data.reactions, this.data.currentUserId) }) },
    toggle(event){ const type=event.currentTarget.dataset.type;const item=this.data.items.find(value=>value.type===type);this.triggerEvent('toggle',{type,reactionId:item&&item.myReactionId||0,reactedByMe:Boolean(item&&item.reactedByMe)}) },
    text(){ this.triggerEvent('text') },
  },
})
