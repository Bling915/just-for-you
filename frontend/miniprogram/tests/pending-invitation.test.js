const test = require('node:test')
const assert = require('node:assert/strict')
const { createPendingInvitation } = require('../utils/pending-invitation')

function memoryStorage() {
  const values = new Map()
  return {
    get: (key) => values.get(key),
    set: (key, value) => values.set(key, value),
    remove: (key) => values.delete(key),
  }
}

test('captures and normalizes an invitation across login redirects', () => {
  const pending = createPendingInvitation(memoryStorage())

  assert.equal(pending.capture(' ab3k9z '), 'AB3K9Z')
  assert.equal(pending.peek(), 'AB3K9Z')
})

test('ignores malformed invitation values', () => {
  const pending = createPendingInvitation(memoryStorage())

  assert.equal(pending.capture('bad!'), '')
  assert.equal(pending.peek(), '')
})

test('consumes a pending invitation only after successful binding', () => {
  const pending = createPendingInvitation(memoryStorage())
  pending.capture('AB3K9Z')

  assert.equal(pending.consume(), 'AB3K9Z')
  assert.equal(pending.peek(), '')
})
