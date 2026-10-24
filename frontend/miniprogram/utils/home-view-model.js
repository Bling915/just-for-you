const PRESENCE_LABELS={in_class:'在上课',miss_you:'想你了',tired:'有点累'}
function statusMoment(presence,latest){if(presence){const status=presence.preset||'';return{status,displayStatus:PRESENCE_LABELS[status]||presence.customText||status,text:presence.customText||''}}if(!latest||typeof latest!=='object')return{};return{...latest,displayStatus:latest.status||''}}
function normalizeHighlights(value){return Array.isArray(value)?value.filter(item=>item&&typeof item==='object'&&item.id!=null):[]}
module.exports={statusMoment,normalizeHighlights}
