function normalize(file,index){return{id:file.id||`${Date.now()}-${index}-${file.path}`,path:file.path,size:file.size||0,contentType:file.contentType||'image/jpeg',thumb:file.thumb||file.path,status:'pending',progress:0,objectKey:'',error:'',errorCode:''}}
function createUploadQueue({requestGrant,putFile,compress=async path=>path,concurrency=2,onChange=()=>{}}){
 let items=[];const notify=()=>onChange(items.map(item=>({...item})));const snapshot=()=>items.map(item=>({...item}))
 function add(files){if(items.length+files.length>9)throw new Error('最多选择 9 张照片');items=items.concat(files.map((file,index)=>normalize(file,items.length+index)));notify();return snapshot()}
 async function upload(item){try{item.status='compressing';item.progress=0;item.error='';item.errorCode='';notify();const uploadPath=await compress(item.path);item.status='granting';notify();const grant=await requestGrant({contentType:item.contentType,size:item.size,path:uploadPath});item.status='uploading';notify();await putFile({filePath:uploadPath,uploadURL:grant.uploadUrl,contentType:grant.contentType||item.contentType,onProgress:progress=>{item.progress=Math.max(0,Math.min(100,Number(progress)||0));notify()},onStateChange:()=>{}});item.status='success';item.progress=100;item.objectKey=grant.objectKey;notify()}catch(error){item.status='failed';item.error=(error&&error.message)||'上传失败';item.errorCode=(error&&error.code)||'UPLOAD_FAILED';notify()}}
 async function run(pending){let cursor=0;const worker=async()=>{while(cursor<pending.length){const item=pending[cursor++];await upload(item)}};await Promise.all(Array.from({length:Math.min(concurrency,pending.length)},worker));return snapshot()}
 async function start(){return run(items.filter(item=>item.status==='pending'))}
 async function retry(index){const item=items[Number(index)];if(!item||item.status!=='failed')return snapshot();item.status='pending';notify();return run([item])}
 async function retryFailed(){items.forEach(item=>{if(item.status==='failed')item.status='pending'});notify();return start()}
 return{add,start,retry,retryFailed,snapshot}
}
module.exports={createUploadQueue}
