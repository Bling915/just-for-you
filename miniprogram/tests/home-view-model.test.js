const test=require('node:test');const assert=require('node:assert/strict')
const {statusMoment}=require('../utils/home-view-model')
test('status orb always receives an object when no partner status exists',()=>{assert.deepEqual(statusMoment(null,null),{})})
test('live presence takes precedence over latest moment status',()=>{assert.deepEqual(statusMoment({preset:'tired',customText:'有点困'},{status:'miss_you'}),{status:'tired',text:'有点困'})})
