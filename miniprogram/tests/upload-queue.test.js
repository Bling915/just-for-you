const test = require('node:test')
const assert = require('node:assert/strict')
const { createUploadQueue } = require('../utils/upload-queue')

const files = (count) => Array.from({length:count},(_,index)=>({path:`/${index}.jpg`,size:10,contentType:'image/jpeg'}))

test('accepts up to nine files and rejects the tenth', () => {
  const queue=createUploadQueue({requestGrant:async()=>({}),putFile:async()=>{},concurrency:2})
  assert.equal(queue.add([]).length,0)
  assert.equal(queue.add(files(9)).length,9)
  assert.throws(()=>queue.add(files(1)),/最多选择 9 张照片/)
})

test('limits concurrency, preserves content type and retries only failed files', async () => {
  let active=0,maxActive=0; const attempts=new Map(); const puts=[]
  const queue=createUploadQueue({concurrency:2,
    requestGrant:async(file)=>({objectKey:`key-${file.path}`,uploadUrl:'https://upload.test',contentType:file.contentType}),
    putFile:async(input)=>{ active++;maxActive=Math.max(maxActive,active); puts.push(input.contentType); const count=(attempts.get(input.filePath)||0)+1;attempts.set(input.filePath,count);await new Promise(r=>setTimeout(r,4));active--;if(input.filePath==='/1.jpg'&&count===1)throw new Error('fail') },
  })
  queue.add(files(3)); await queue.start()
  assert.equal(maxActive,2);assert.deepEqual(puts,['image/jpeg','image/jpeg','image/jpeg'])
  assert.deepEqual(queue.snapshot().map(x=>x.status),['success','failed','success'])
  await queue.retryFailed()
  assert.deepEqual(queue.snapshot().map(x=>x.status),['success','success','success'])
  assert.equal(attempts.get('/0.jpg'),1);assert.equal(attempts.get('/1.jpg'),2);assert.equal(attempts.get('/2.jpg'),1)
})

test('reports compression and upload progress and retries one failed file',async()=>{
  const attempts=new Map()
  const queue=createUploadQueue({
    requestGrant:async(file)=>({objectKey:`key-${file.path}`,uploadUrl:'https://upload.test',contentType:file.contentType}),
    compress:async(path)=>`/compressed${path}`,
    putFile:async({filePath,onProgress})=>{onProgress(35);onProgress(100);const count=(attempts.get(filePath)||0)+1;attempts.set(filePath,count);if(filePath==='/compressed/1.jpg'&&count===1){const error=new Error('network');error.code='NETWORK';throw error}},
  })
  queue.add(files(2));await queue.start()
  assert.deepEqual(queue.snapshot().map(item=>({status:item.status,progress:item.progress,errorCode:item.errorCode})),[
    {status:'success',progress:100,errorCode:''},{status:'failed',progress:100,errorCode:'NETWORK'},
  ])
  await queue.retry(1)
  assert.deepEqual(queue.snapshot().map(item=>item.status),['success','success'])
  assert.equal(attempts.get('/compressed/0.jpg'),1)
})
