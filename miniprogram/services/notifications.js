const {http}=require('./http-client')
function normalizeSubscriptionResult(templateIds,result={}){return templateIds.reduce((out,id)=>{out[id]=result[id]==='accept';return out},{})}
function requestSubscriptions(templateIds){const ids=templateIds.filter(Boolean);if(!ids.length)return Promise.resolve({});return new Promise((resolve,reject)=>wx.requestSubscribeMessage({tmplIds:ids,success:r=>resolve(normalizeSubscriptionResult(ids,r)),fail:reject}))}
const notificationService={get:()=>http.get('/v1/notification-preferences'),update:value=>http.put('/v1/notification-preferences',value)}
module.exports={notificationService,normalizeSubscriptionResult,requestSubscriptions}
