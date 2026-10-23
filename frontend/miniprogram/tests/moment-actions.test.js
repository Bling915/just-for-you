const test = require('node:test')
const assert = require('node:assert/strict')
const { reactionView, toggleReaction, canReactToMoment, canShareToCouple, normalizeReactionText } = require('../utils/moment-actions')

test('reaction view aggregates counts and marks my reaction', () => {
  const reactions = [{ id: 1, type: 'hug', userId: 2 }, { id: 2, type: 'hug', userId: 1 }, { id: 3, type: 'seen', userId: 2 }]
  assert.deepEqual(reactionView(reactions, 1).map(({ type, count, reactedByMe }) => ({ type, count, reactedByMe })), [
    { type: 'seen', count: 1, reactedByMe: false },
    { type: 'hug', count: 2, reactedByMe: true },
  ])
})

test('toggle reaction is optimistic and returns rollback state', () => {
  const moment = { id: 7, reactions: [{ id: 4, type: 'hug', userId: 2 }] }
  const { next, rollback } = toggleReaction(moment, 'hug', 1)
  assert.equal(next.reactions.length, 2)
  assert.deepEqual(rollback, moment)
  assert.notEqual(rollback, next)
})

test('only the other partner can react to a moment', () => {
  assert.equal(canReactToMoment({ authorId: 1 }, 1), false)
  assert.equal(canReactToMoment({ authorId: 2 }, 1), true)
  assert.equal(canReactToMoment(null, 1), false)
})

test('only personal author can share and text uses unicode 120 limit', () => {
  assert.equal(canShareToCouple({ authorId: 1, spaceType: 'personal', kind: 'standard' }, 1), true)
  assert.equal(canShareToCouple({ authorId: 2, spaceType: 'personal', kind: 'standard' }, 1), false)
  assert.equal(normalizeReactionText('  想你  '), '想你')
  assert.throws(() => normalizeReactionText('🌿'.repeat(121)), /120/)
})
