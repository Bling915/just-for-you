const { countRunes,truncateRunes }=require('../../utils/runes')
const { createUploadQueue }=require('../../utils/upload-queue')
const { compressSelectedImage }=require('../../utils/image-compression')
const { uploadService }=require('../../services/uploads')
const { momentService }=require('../../services/moments')
const { draftStore }=require('../../store/draft')
const { sessionStore }=require('../../store/session')
const { relationshipService }=require('../../services/relationships')

const pad=value=>String(value).padStart(2,'0')
function localParts(value){const date=value?new Date(value):new Date();return{date:`${date.getFullYear()}-${pad(date.getMonth()+1)}-${pad(date.getDate())}`,time:`${pad(date.getHours())}:${pad(date.getMinutes())}`}}
function iso(date,time){return new Date(`${date}T${time}:00`).toISOString()}

Page({
  data:{text:'',status:'',placeName:'',latitude:null,longitude:null,files:[],occurredAt:'',occurredDate:'',occurredTime:'',today:'',textCount:0,placeCount:0,publishing:false,completed:0,total:0,statuses:[{code:'in_class',label:'在上课',icon:'✎'},{code:'miss_you',label:'想你了',icon:'♡'},{code:'tired',label:'有点累',icon:'☾'}]},
  onLoad(){const draft=draftStore.load();const parts=localParts(draft.occurredAt);const today=localParts().date;this.setData({...draft,...parts,occurredDate:parts.date,occurredTime:parts.time,today,textCount:countRunes(draft.text),placeCount:countRunes(draft.placeName)})},
  async onShow(){if(!sessionStore.snapshot().accessToken)return wx.reLaunch({url:'/pages/welcome/index'});try{await relationshipService.getCurrent()}catch(error){if(error.code==='NO_ACTIVE_COUPLE'){draftStore.clear();return wx.reLaunch({url:'/pages/bind/index'})}}},
  onUnload(){if(!this.data.publishing)this.save()},
  inputText(event){const text=truncateRunes(event.detail.value,500);this.setData({text,textCount:countRunes(text)});this.save()},
  inputPlace(event){const placeName=truncateRunes(event.detail.value,80);this.setData({placeName,placeCount:countRunes(placeName)});this.save()},
  selectStatus(event){const code=event.currentTarget.dataset.code;this.setData({status:this.data.status===code?'':code});this.save()},
  changeDate(event){this.setData({occurredDate:event.detail.value});this.save()},changeTime(event){this.setData({occurredTime:event.detail.value});this.save()},
  chooseLocation(){wx.chooseLocation({success:place=>{this.setData({placeName:truncateRunes(place.name||place.address||'',80),placeCount:countRunes(place.name||place.address||''),latitude:place.latitude,longitude:place.longitude});this.save()}})},
  clearLocation(){this.setData({placeName:'',placeCount:0,latitude:null,longitude:null});this.save()},
  filesChanged(event){this.queue=null;const files=event.detail.files.map(({path,thumb,size,contentType})=>({path,thumb,size,contentType}));this.setData({files});this.save()},
  async retryPhoto(event){if(!this.queue||this.data.publishing)return;this.setData({publishing:true});const result=await this.queue.retry(event.detail.index);if(result.some(item=>item.status==='failed')){this.setData({publishing:false});return wx.showToast({title:'还是没传好，再试一次吧',icon:'none'})}await this.completePublish(result)},
  save(){draftStore.patch({text:this.data.text,status:this.data.status,placeName:this.data.placeName,latitude:this.data.latitude,longitude:this.data.longitude,occurredAt:iso(this.data.occurredDate,this.data.occurredTime),files:this.data.files.map(({path,thumb,size,contentType})=>({path,thumb,size,contentType}))})},
  payload(items){return{text:this.data.text.trim(),status:this.data.status,placeName:this.data.placeName.trim(),latitude:this.data.latitude,longitude:this.data.longitude,occurredAt:iso(this.data.occurredDate,this.data.occurredTime),photoKeys:items.map(item=>item.objectKey)}},
  async completePublish(items){try{await momentService.publish(this.payload(items));draftStore.clear();this.queue=null;this.setData({publishing:false});wx.showToast({title:'小事记下了',icon:'success'});setTimeout(()=>wx.navigateBack(),500)}catch(_){this.setData({publishing:false});this.save();wx.showToast({title:'发布失败，草稿已保存',icon:'none'})}},
  async publish(){
    if(this.data.publishing)return
    if(!this.data.text.trim()&&!this.data.status&&!this.data.placeName.trim()&&!this.data.files.length)return wx.showToast({title:'先记下一点什么吧',icon:'none'})
    this.setData({publishing:true,total:this.data.files.length,completed:0})
    const queue=createUploadQueue({requestGrant:uploadService.requestGrant,putFile:uploadService.putFile,compress:compressSelectedImage,concurrency:2,onChange:items=>this.setData({files:items,completed:items.filter(item=>item.status==='success').length})});this.queue=queue
    try{queue.add(this.data.files);await queue.start();const result=queue.snapshot();if(result.some(item=>item.status==='failed')){this.setData({publishing:false});this.save();return wx.showToast({title:'有照片没传好，点照片重试',icon:'none'})}await this.completePublish(result)}catch(_){this.setData({publishing:false});this.save();wx.showToast({title:'发布失败，草稿已保存',icon:'none'})}
  },
})
