function statusMoment(presence,latest){if(presence)return{status:presence.preset||'',text:presence.customText||''};return latest&&typeof latest==='object'?latest:{}}
module.exports={statusMoment}
