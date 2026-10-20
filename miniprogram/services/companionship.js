const {http}=require('./http-client')
function createCompanionshipService(client){return{
  getSnapshot:()=>client.get('/v1/presence'),
  updatePresence(input){return client.put('/v1/presence',{preset:String(input.preset||'').trim(),customText:String(input.customText||'').trim(),expiresAt:input.expiresAt||null})},
  sendNudge(kind){return client.post('/v1/nudges',{kind:String(kind||'').trim()})},
}}
function effectivePresence(snapshot,now=new Date()){const presence=snapshot&&snapshot.partnerPresence;if(!presence)return null;if(presence.expiresAt&&new Date(presence.expiresAt).getTime()<=now.getTime())return null;return presence.customText||presence.preset||null}
const companionshipService=createCompanionshipService(http)
module.exports={createCompanionshipService,companionshipService,effectivePresence}
