const { http } = require('./http-client')

const momentService = {
  getHome: () => http.get('/v1/home'),
  getTimeline({ space = 'shared', owner = '', cursor = '', limit = 20 } = {}) {
    const query = [`space=${encodeURIComponent(space)}`, `limit=${encodeURIComponent(limit)}`]
    if (owner) query.push(`owner=${encodeURIComponent(owner)}`)
    if (cursor) query.push(`cursor=${encodeURIComponent(cursor)}`)
    return http.get(`/v1/moments?${query.join('&')}`)
  },
  publish: (input) => http.post('/v1/moments', input),
  updateMoment: (id, input) => http.patch(`/v1/moments/${encodeURIComponent(id)}`, input),
  deleteMoment: (id) => http.delete(`/v1/moments/${encodeURIComponent(id)}`),
  shareToCouple: (id, input = {}) => http.post(`/v1/moments/${encodeURIComponent(id)}/share-to-couple`, input),
  react: (id, input) => http.post(`/v1/moments/${encodeURIComponent(id)}/reactions`, input),
  deleteReaction: (momentId, reactionId) => http.delete(`/v1/moments/${encodeURIComponent(momentId)}/reactions/${encodeURIComponent(reactionId)}`),
  savePerspective: (momentId, input) => http.put(`/v1/moments/${encodeURIComponent(momentId)}/perspective`, input),
  deletePerspective: (momentId) => http.delete(`/v1/moments/${encodeURIComponent(momentId)}/perspective`),
  getCalendar: month => http.get(`/v1/moments/calendar?month=${encodeURIComponent(month)}`),
  getMap: ({from,to}) => http.get(`/v1/moments/map?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),
  getMonthlyReport: month => http.get(`/v1/reports/monthly?month=${encodeURIComponent(month)}`),
  getResurfacedMemory: () => http.get('/v1/memories/resurface'),
}

module.exports = { momentService }
