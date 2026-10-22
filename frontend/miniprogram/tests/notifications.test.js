const test=require('node:test');const assert=require('node:assert/strict');const {normalizeSubscriptionResult}=require('../services/notifications')
test('subscription accept and reject results normalize to booleans',()=>{assert.deepEqual(normalizeSubscriptionResult(['a','b','c'],{a:'accept',b:'reject',c:'ban'}),{a:true,b:false,c:false})})
