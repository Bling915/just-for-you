function normalizeSpaceSelection(input = {}) {
  if (input.spaceType === 'personal') {
    return {
      spaceType: 'personal',
      visibility: input.visibility === 'partner' ? 'partner' : 'private',
    }
  }
  return { spaceType: 'shared', visibility: 'couple' }
}

function publishButtonLabel(input = {}) {
  const selection = normalizeSpaceSelection(input)
  if (selection.spaceType === 'shared') return '发布到我们的小事'
  return selection.visibility === 'partner' ? '分享给 TA' : '仅自己保存'
}

function timelineQueryKey(input = {}) {
  if (input.space === 'personal') return `personal:${input.owner === 'partner' ? 'partner' : 'me'}`
  return 'shared'
}

module.exports = { normalizeSpaceSelection, publishButtonLabel, timelineQueryKey }
