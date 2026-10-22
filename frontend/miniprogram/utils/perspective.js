function canAddPerspective(moment,currentUserId){return Boolean(moment&&moment.spaceType==='shared'&&Number(moment.authorId)!==Number(currentUserId)&&!moment.perspective)}
function buildPerspectivePayload(text,files=[]){const value=String(text||'').trim();if(files.length>3)throw new Error('最多选择 3 张照片');if(!value&&!files.length)throw new Error('请写点内容或选择照片');if(Array.from(value).length>500)throw new Error('内容不能超过 500 个字');return{text:value,photoKeys:files.map(file=>file.objectKey).filter(Boolean)}}
module.exports={canAddPerspective,buildPerspectivePayload}
