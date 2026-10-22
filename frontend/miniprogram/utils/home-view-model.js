function statusMoment(presence,latest){if(presence)return{status:presence.preset||'',text:presence.customText||''};return latest&&typeof latest==='object'?latest:{}}
function normalizeHighlights(value){return Array.isArray(value)?value.filter(item=>item&&typeof item==='object'&&item.id!=null):[]}
module.exports={statusMoment,normalizeHighlights}
