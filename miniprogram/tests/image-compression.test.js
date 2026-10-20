const test=require('node:test')
const assert=require('node:assert/strict')
const {compressImage}=require('../utils/image-compression')

test('uses compressed image path when compression succeeds',async()=>{
  const wxAPI={compressImage:({success})=>success({tempFilePath:'/compressed.jpg'})}
  assert.equal(await compressImage(wxAPI,'/original.jpg'),'/compressed.jpg')
})

test('falls back to original image when compression is unavailable',async()=>{
  const wxAPI={compressImage:({fail})=>fail(new Error('unsupported'))}
  assert.equal(await compressImage(wxAPI,'/original.jpg'),'/original.jpg')
})
