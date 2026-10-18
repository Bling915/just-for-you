const test=require('node:test');const assert=require('node:assert/strict');const {imageContentType}=require('../utils/media-type')
test('chooseMedia image category is mapped to a supported MIME by extension',()=>{assert.equal(imageContentType('/tmp/a.jpg','image'),'image/jpeg');assert.equal(imageContentType('/tmp/a.PNG','image'),'image/png');assert.equal(imageContentType('/tmp/a.webp','image'),'image/webp')})
test('unsupported image formats are rejected',()=>assert.throws(()=>imageContentType('/tmp/a.gif','image'),/仅支持/))
