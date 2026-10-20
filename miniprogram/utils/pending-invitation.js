const { storage } = require('./storage')

const KEY = 'pendingInvitationCode'
const INVITATION_PATTERN = /^[23456789ABCDEFGHJKLMNPQRSTUVWXYZ]{6}$/

function normalize(value) {
  const code = String(value || '').trim().toUpperCase()
  return INVITATION_PATTERN.test(code) ? code : ''
}

function createPendingInvitation(targetStorage) {
  return {
    capture(value) {
      const code = normalize(value)
      if (code) targetStorage.set(KEY, code)
      return code
    },
    peek() { return normalize(targetStorage.get(KEY)) },
    consume() {
      const code = normalize(targetStorage.get(KEY))
      targetStorage.remove(KEY)
      return code
    },
    clear() { targetStorage.remove(KEY) },
  }
}

const pendingInvitation = createPendingInvitation(storage)

module.exports = { createPendingInvitation, pendingInvitation }
