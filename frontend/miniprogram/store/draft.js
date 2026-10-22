const { storage: wxStorage } = require('../utils/storage')
const KEY='momentDraft'; const EMPTY={text:'',status:'',placeName:'',latitude:null,longitude:null,occurredAt:'',spaceType:'shared',visibility:'couple',files:[]}
function createDraftStore(storage){let state={...EMPTY};return{load(){state={...EMPTY,...(storage.get(KEY)||{})};return this.snapshot()},patch(input){state={...state,...input};storage.set(KEY,state);return this.snapshot()},clear(){state={...EMPTY};storage.remove(KEY)},snapshot(){return {...state,files:(state.files||[]).map(file=>({...file}))}}}}
const draftStore=createDraftStore(wxStorage)
module.exports={createDraftStore,draftStore}
