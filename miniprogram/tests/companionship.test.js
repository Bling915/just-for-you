const test=require('node:test')
const assert=require('node:assert/strict')
const {createCompanionshipService,effectivePresence}=require('../services/companionship')

test('normalizes presence and nudge request payloads',async()=>{
  const calls=[]
  const client={put:async(path,body)=>{calls.push({method:'PUT',path,body});return body},post:async(path,body)=>{calls.push({method:'POST',path,body});return body},get:async()=>({})}
  const service=createCompanionshipService(client)
  await service.updatePresence({preset:' custom ',customText:' 去图书馆啦 ',expiresAt:''})
  await service.sendNudge(' hug ')
  assert.deepEqual(calls,[
    {method:'PUT',path:'/v1/presence',body:{preset:'custom',customText:'去图书馆啦',expiresAt:null}},
    {method:'POST',path:'/v1/nudges',body:{kind:'hug'}},
  ])
})

test('expired partner presence is hidden from the view',()=>{
  const snapshot={partnerPresence:{preset:'tired',expiresAt:'2026-10-20T11:59:59Z'}}
  assert.equal(effectivePresence(snapshot,new Date('2026-10-20T12:00:00Z')),null)
  assert.equal(effectivePresence({partnerPresence:{preset:'in_class',expiresAt:null}},new Date()),'in_class')
})
