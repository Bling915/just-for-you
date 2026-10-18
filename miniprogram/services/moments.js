const { http } = require('./http-client')

const momentService = {
  getHome: () => http.get('/v1/home'),
  getTimeline({ cursor = '', limit = 20 } = {}) {
    const query = [`limit=${encodeURIComponent(limit)}`]
    if (cursor) query.push(`cursor=${encodeURIComponent(cursor)}`)
    return http.get(`/v1/moments?${query.join('&')}`)
  },
  publish: (input) => http.post('/v1/moments', input),
  deleteMoment: (id) => http.delete(`/v1/moments/${encodeURIComponent(id)}`),
}

module.exports = { momentService }
