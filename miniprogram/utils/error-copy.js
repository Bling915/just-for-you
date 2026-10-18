const COPY = Object.freeze({
  INVITE_INVALID: '这个邀请码无效，请检查后重新输入',
  INVITE_EXPIRED: '邀请码已经过期，请让对方重新创建',
  INVITE_USED: '邀请码已经使用过，请让对方重新创建',
  CANNOT_BIND_SELF: '不能和自己绑定，请使用对方的邀请码',
  ALREADY_BOUND: '你或对方已经有有效的情侣关系',
  NO_ACTIVE_COUPLE: '还没有绑定伴侣，一起创建新的关系吧',
  TEMPORARY_UNAVAILABLE: '服务暂时不可用，请稍后重试',
  NETWORK_ERROR: '网络连接失败，请检查网络后重试',
});

function messageForError(code) {
  return COPY[code] || COPY.TEMPORARY_UNAVAILABLE;
}

function normalizeInvitationCode(value) {
  return String(value || '').replace(/\s+/g, '').toUpperCase().slice(0, 6);
}

module.exports = { messageForError, normalizeInvitationCode };
