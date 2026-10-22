const test = require('node:test')
const assert = require('node:assert/strict')
const { formatRelativeTime } = require('../utils/date')

const now = new Date('2026-10-18T15:00:00+08:00')

test('formats recent, same-day and yesterday moments', () => {
  assert.equal(formatRelativeTime('2026-10-18T14:59:40+08:00', now), '刚刚')
  assert.equal(formatRelativeTime('2026-10-18T14:48:00+08:00', now), '12 分钟前')
  assert.equal(formatRelativeTime('2026-10-18T14:20:00+08:00', now), '今天 14:20')
  assert.equal(formatRelativeTime('2026-10-17T22:10:00+08:00', now), '昨天 22:10')
})

test('formats dates from another year and tolerates invalid input', () => {
  assert.equal(formatRelativeTime('2025-12-31T23:30:00+08:00', now), '2025年12月31日 23:30')
  assert.equal(formatRelativeTime('not-a-date', now), '')
  assert.equal(formatRelativeTime('', now), '')
})
