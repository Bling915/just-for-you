const { http } = require('./http-client');

function createRelationshipService(client) {
  return {
    createInvitation: () => client.post('/v1/couples/invitations', {}),
    previewInvitation: (code) => client.get(`/v1/couples/invitations/${encodeURIComponent(code)}`),
    bind: (code) => client.post('/v1/couples/bind', { code }),
    enableSoloTest: () => client.post('/v1/couples/solo-test', {}),
    getCurrent: () => client.get('/v1/couples/me'),
    end: () => client.delete('/v1/couples/me'),
  };
}

const relationshipService = createRelationshipService(http);

module.exports = { createRelationshipService, relationshipService };
