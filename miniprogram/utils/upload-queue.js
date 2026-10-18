function normalize(file,index){ return {id:file.id||`${Date.now()}-${index}-${file.path}`,path:file.path,size:file.size||0,contentType:file.contentType||'image/jpeg',thumb:file.thumb||file.path,status:'pending',objectKey:'',error:''} }

function createUploadQueue({requestGrant,putFile,concurrency=2,onChange=()=>{}}){
  let items=[]
  const notify=()=>onChange(items.map(item=>({...item})))
  function snapshot(){return items.map(item=>({...item}))}
  function add(files){if(items.length+files.length>9)throw new Error('最多选择 9 张照片');items=items.concat(files.map((file,index)=>normalize(file,items.length+index)));notify();return snapshot()}
  async function upload(item){
    try{item.status='granting';item.error='';notify();const grant=await requestGrant({contentType:item.contentType,size:item.size,path:item.path});item.status='uploading';notify();await putFile({filePath:item.path,uploadURL:grant.uploadUrl,contentType:grant.contentType||item.contentType,onStateChange:()=>{}});item.status='success';item.objectKey=grant.objectKey;notify()}
    catch(error){item.status='failed';item.error=(error&&error.message)||'上传失败';notify()}
  }
  async function start(){const pending=items.filter(item=>item.status==='pending');let cursor=0;const worker=async()=>{while(cursor<pending.length){const item=pending[cursor++];await upload(item)}};await Promise.all(Array.from({length:Math.min(concurrency,pending.length)},worker));return snapshot()}
  async function retryFailed(){items.forEach(item=>{if(item.status==='failed')item.status='pending'});notify();return start()}
  return {add,start,retryFailed,snapshot}
}

module.exports={createUploadQueue}
