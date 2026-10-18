const { http } = require('./http-client');

const relationshipService = {
  createInvitation: () => http.post('/v1/couples/invitations', {}),
  previewInvitation: (code) => http.get(`/v1/couples/invitations/${encodeURIComponent(code)}`),
  bind: (code) => http.post('/v1/couples/bind', { code }),
  getCurrent: () => http.get('/v1/couples/me'),
  end: () => http.delete('/v1/couples/me'),
};

module.exports = { relationshipService };
