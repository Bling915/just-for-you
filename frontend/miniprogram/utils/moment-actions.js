const ORDER = ['seen', 'hug', 'together', 'collect']
const LABELS = { seen: '看见了', hug: '抱抱', together: '一起', collect: '珍藏' }

function reactionView(reactions = [], currentUserId = 0) {
  const groups = new Map()
  reactions.filter(item => item.type !== 'text').forEach(item => {
    const value = groups.get(item.type) || { type: item.type, label: LABELS[item.type] || item.type, count: 0, reactedByMe: false, myReactionId: 0 }
    value.count += 1
    if (Number(item.userId) === Number(currentUserId)) { value.reactedByMe = true; value.myReactionId = item.id }
    groups.set(item.type, value)
  })
  return ORDER.filter(type => groups.has(type)).map(type => groups.get(type))
}

function toggleReaction(moment, type, currentUserId) {
  const rollback = { ...moment, reactions: (moment.reactions || []).map(item => ({ ...item })) }
  const own = rollback.reactions.find(item => item.type === type && Number(item.userId) === Number(currentUserId))
  const reactions = own
    ? rollback.reactions.filter(item => item !== own)
    : rollback.reactions.concat({ id: `pending-${type}`, type, userId: currentUserId, reactedByMe: true, pending: true })
  return { next: { ...moment, reactions }, rollback, existing: own || null }
}

function canShareToCouple(moment, currentUserId) {
  return Boolean(moment && Number(moment.authorId) === Number(currentUserId) && moment.spaceType === 'personal' && (!moment.kind || moment.kind === 'standard'))
}

function canReactToMoment(moment, currentUserId) {
  return Boolean(moment && Number(moment.authorId) !== Number(currentUserId))
}

function normalizeReactionText(value) {
  const text = String(value || '').trim()
  if (!text) throw new Error('请输入回应')
  if (Array.from(text).length > 120) throw new Error('回应不能超过 120 个字')
  return text
}

module.exports = { reactionView, toggleReaction, canReactToMoment, canShareToCouple, normalizeReactionText }
