function pad(value) { return String(value).padStart(2, '0') }
function clock(date) { return `${pad(date.getHours())}:${pad(date.getMinutes())}` }
function dayStart(date) { return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime() }

function formatRelativeTime(iso, nowValue = new Date()) {
  if (!iso) return ''
  const date = new Date(iso)
  const now = new Date(nowValue)
  if (Number.isNaN(date.getTime()) || Number.isNaN(now.getTime())) return ''
  const difference = now.getTime() - date.getTime()
  if (difference >= 0 && difference < 60000) return '刚刚'
  if (difference >= 60000 && difference < 30 * 60000) return `${Math.floor(difference / 60000)} 分钟前`
  const days = Math.round((dayStart(now) - dayStart(date)) / 86400000)
  if (days === 0) return `今天 ${clock(date)}`
  if (days === 1) return `昨天 ${clock(date)}`
  if (date.getFullYear() === now.getFullYear()) return `${date.getMonth() + 1}月${date.getDate()}日 ${clock(date)}`
  return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日 ${clock(date)}`
}

module.exports = { formatRelativeTime }
