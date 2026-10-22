const { momentService } = require('../services/moments')
const { toggleReaction, normalizeReactionText } = require('./moment-actions')

async function reactOptimistically(event, currentUserId, update) {
  const { moment, type } = event.detail
  const change = toggleReaction(moment, type, currentUserId)
  update(change.next)
  try {
    if (change.existing) await momentService.deleteReaction(moment.id, change.existing.id)
    else {
      const saved = await momentService.react(moment.id, { type })
      update({ ...change.next, reactions: change.next.reactions.map(item => item.pending && item.type === type ? saved : item) })
    }
  } catch (_) {
    update(change.rollback)
    wx.showToast({ title: '回应没有送达，再试一次吧', icon: 'none' })
  }
}

function promptTextReaction(event, update) {
  const moment = event.detail.moment
  wx.showModal({ title: '写给 TA 的一句话', editable: true, placeholderText: '最多 120 个字', success: async result => {
    if (!result.confirm) return
    try {
      const text = normalizeReactionText(result.content)
      const saved = await momentService.react(moment.id, { type: 'text', text })
      update({ ...moment, reactions: (moment.reactions || []).filter(item => !(item.type === 'text' && item.reactedByMe)).concat(saved) })
    } catch (error) { wx.showToast({ title: error.message || '发送失败，请稍后再试', icon: 'none' }) }
  } })
}

function shareToCouple(event, onSuccess) {
  const moment = event.detail.moment
  wx.showModal({ title: '分享给我们', editable: true, placeholderText: '可以补一句想说的话（选填）', confirmText: '分享', success: async result => {
    if (!result.confirm) return
    try { await momentService.shareToCouple(moment.id, { message: String(result.content || '').trim() }); wx.showToast({ title: '已放进共同时间线', icon: 'none' }); if (onSuccess) onSuccess() }
    catch (_) { wx.showToast({ title: '分享失败，请稍后再试', icon: 'none' }) }
  } })
}

module.exports = { reactOptimistically, promptTextReaction, shareToCouple }
