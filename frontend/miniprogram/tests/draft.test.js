const test = require('node:test')
const assert = require('node:assert/strict')
const { createDraftStore } = require('../store/draft')

test('draft survives store recreation and clear removes it', () => {
  const memory={}; const storage={get:k=>memory[k],set:(k,v)=>memory[k]=v,remove:k=>delete memory[k]}
  createDraftStore(storage).patch({text:'晚风很好',status:'想你了',placeName:'操场',files:[{path:'/a.jpg'}]})
  const restored=createDraftStore(storage); restored.load()
  assert.deepEqual(restored.snapshot(),{text:'晚风很好',status:'想你了',placeName:'操场',latitude:null,longitude:null,occurredAt:'',spaceType:'shared',visibility:'couple',files:[{path:'/a.jpg'}]})
  restored.clear();assert.deepEqual(restored.snapshot(),{text:'',status:'',placeName:'',latitude:null,longitude:null,occurredAt:'',spaceType:'shared',visibility:'couple',files:[]})
})
