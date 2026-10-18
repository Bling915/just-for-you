const test = require('node:test');
const assert = require('node:assert/strict');

const { messageForError, normalizeInvitationCode } = require('../utils/error-copy');

test('relationship errors have actionable Chinese messages', () => {
  const codes = ['INVITE_INVALID', 'INVITE_EXPIRED', 'INVITE_USED', 'CANNOT_BIND_SELF', 'ALREADY_BOUND', 'NO_ACTIVE_COUPLE'];
  for (const code of codes) {
    const message = messageForError(code);
    assert.ok(message.length >= 6, `${code} should have useful copy`);
    assert.equal(/[a-z]{3,}/i.test(message), false);
  }
});

test('temporary and unknown errors use retry copy', () => {
  assert.equal(messageForError('TEMPORARY_UNAVAILABLE'), '服务暂时不可用，请稍后重试');
  assert.equal(messageForError('SOME_INTERNAL_FAILURE'), '服务暂时不可用，请稍后重试');
});

test('invitation code is uppercase, trimmed and limited to six characters', () => {
  assert.equal(normalizeInvitationCode(' ab 3k9z extra '), 'AB3K9Z');
});
