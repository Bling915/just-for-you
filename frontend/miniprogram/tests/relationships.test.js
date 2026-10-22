const test = require('node:test');
const assert = require('node:assert/strict');
const { createRelationshipService } = require('../services/relationships');

test('solo test mode uses the protected relationship endpoint', async () => {
  const calls = [];
  const service = createRelationshipService({
    post: async (path, body) => {
      calls.push({ path, body });
      return { id: 12, userAId: 7, userBId: 1007 };
    },
  });

  const relationship = await service.enableSoloTest();

  assert.deepEqual(calls, [{ path: '/v1/couples/solo-test', body: {} }]);
  assert.equal(relationship.id, 12);
});
