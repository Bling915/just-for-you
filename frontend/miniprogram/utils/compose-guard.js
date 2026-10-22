const { countRunes } = require('./runes')
const { normalizeSpaceSelection } = require('./space-model')

/**
 * 新增记录页的「离开时要不要留草稿」决策。
 *
 * 原来的实现是 onUnload 里静默保存，用户不知道草稿被存了，
 * 下次进来莫名其妙出现上次的内容。这里改成显式询问。
 *
 * 规则：
 * - 空白页（没字、没状态、没地点、没照片）→ 直接走，不打扰用户
 * - 正在发布中 → 不拦，发布流程自己管草稿
 * - 有内容但未发布 → 弹窗问用户
 */

/**
 * 草稿里只存「跨会话还能可靠恢复」的字段。
 *
 * 照片是 wx.chooseMedia 给的临时文件路径（wxfile://tmp_xxx），
 * 小程序退出后会被系统清理 —— 存进草稿，下次恢复就是一堆裂图。
 * 所以草稿只留文字 / 状态 / 地点 / 时间，照片交给用户重选。
 */
const DRAFT_FIELDS = ['text', 'status', 'placeName', 'latitude', 'longitude', 'occurredAt', 'spaceType', 'visibility']

/** 页面当前是否算「有内容」，用于决定要不要打扰用户 */
function hasContent(form = {}) {
  return Boolean(
    String(form.text || '').trim() ||
      String(form.status || '').trim() ||
      String(form.placeName || '').trim() ||
      (form.latitude !== null && form.latitude !== undefined) ||
      (Array.isArray(form.files) && form.files.length > 0),
  )
}

/**
 * 决定离开页面时该做什么。
 * @returns {{action:'leave'|'confirm'|'blocked', reason?:string}}
 *   leave   —— 直接离开，不用问
 *   confirm —— 需要弹窗问用户
 *   blocked —— 正在发布，不能离开
 */
function decideOnExit(form = {}, { publishing = false } = {}) {
  if (publishing) return { action: 'blocked', reason: '正在发布，请稍候' }
  if (!hasContent(form)) return { action: 'leave', reason: 'empty' }
  return { action: 'confirm', reason: 'has-content' }
}

/** 草稿是否为空 —— 空草稿不值得存 */
function isDraftEmpty(draft = {}) {
  return !hasContent(draft)
}

/**
 * 把页面状态压成可持久化的草稿。
 * 只保留可靠字段，照片一律丢弃（临时路径必失效）。
 */
function buildDraftPayload(form = {}) {
  const payload = {}
  for (const key of DRAFT_FIELDS) {
    const value = form[key]
    payload[key] = value === undefined || value === null ? (key === 'latitude' || key === 'longitude' ? null : '') : value
  }
  payload.files = []
  return payload
}

/**
 * 从草稿恢复出页面表单。
 * 照片一律不恢复（草稿里存的路径早已失效）。
 */
function restoreDraft(draft = {}) {
	const space = normalizeSpaceSelection(draft)
  return {
    text: String(draft.text || ''),
    status: String(draft.status || ''),
    placeName: String(draft.placeName || ''),
    latitude: draft.latitude === undefined ? null : draft.latitude,
    longitude: draft.longitude === undefined ? null : draft.longitude,
    occurredAt: String(draft.occurredAt || ''),
		spaceType: space.spaceType,
		visibility: space.visibility,
    files: [],
  }
}

module.exports = {
  hasContent,
  decideOnExit,
  isDraftEmpty,
  buildDraftPayload,
  restoreDraft,
  countRunes,
}
