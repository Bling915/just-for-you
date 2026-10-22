const test = require('node:test')
const assert = require('node:assert/strict')
const { normalizeSpaceSelection, publishButtonLabel, timelineQueryKey } = require('../utils/space-model')
const { createTimelineStore } = require('../store/timeline')

test('new compose defaults to shared while personal defaults to private', () => {
  assert.deepEqual(normalizeSpaceSelection(), { spaceType: 'shared', visibility: 'couple' })
  assert.deepEqual(normalizeSpaceSelection({ spaceType: 'personal' }), { spaceType: 'personal', visibility: 'private' })
  assert.deepEqual(normalizeSpaceSelection({ spaceType: 'personal', visibility: 'partner' }), { spaceType: 'personal', visibility: 'partner' })
  assert.deepEqual(normalizeSpaceSelection({ spaceType: 'shared', visibility: 'private' }), { spaceType: 'shared', visibility: 'couple' })
})

test('publish labels make the destination and privacy explicit', () => {
  assert.equal(publishButtonLabel({ spaceType: 'shared' }), '发布到我们的小事')
  assert.equal(publishButtonLabel({ spaceType: 'personal', visibility: 'private' }), '仅自己保存')
  assert.equal(publishButtonLabel({ spaceType: 'personal', visibility: 'partner' }), '分享给 TA')
})

test('timeline query keys separate all three feeds', () => {
  assert.equal(timelineQueryKey({ space: 'shared' }), 'shared')
  assert.equal(timelineQueryKey({ space: 'personal', owner: 'me' }), 'personal:me')
  assert.equal(timelineQueryKey({ space: 'personal', owner: 'partner' }), 'personal:partner')
})

test('timeline stores keep filters cursors and results isolated', async () => {
  const calls = []
  const service = {
    async getTimeline(input) {
      calls.push(input)
      const owner = input.owner || 'couple'
      return { moments: [{ id: `${input.space}:${owner}` }], nextCursor: `${owner}-next` }
    },
  }
  const shared = createTimelineStore(service, { space: 'shared' })
  const mine = createTimelineStore(service, { space: 'personal', owner: 'me' })
  const partner = createTimelineStore(service, { space: 'personal', owner: 'partner' })
  await Promise.all([shared.loadFirst(), mine.loadFirst(), partner.loadFirst()])
  assert.deepEqual(calls.map(({ space, owner }) => [space, owner || '']), [
    ['shared', ''], ['personal', 'me'], ['personal', 'partner'],
  ])
  assert.equal(shared.snapshot().moments[0].id, 'shared:couple')
  assert.equal(mine.snapshot().moments[0].id, 'personal:me')
  assert.equal(partner.snapshot().moments[0].id, 'personal:partner')
  assert.notEqual(mine.snapshot().cursor, partner.snapshot().cursor)
})
