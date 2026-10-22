const { countRunes,truncateRunes }=require('../../utils/runes')
const { createUploadQueue }=require('../../utils/upload-queue')
const { compressSelectedImage }=require('../../utils/image-compression')
const { uploadService }=require('../../services/uploads')
const { momentService }=require('../../services/moments')
const { draftStore }=require('../../store/draft')
const { decideOnExit, isDraftEmpty, buildDraftPayload, restoreDraft }=require('../../utils/compose-guard')
const { sessionStore }=require('../../store/session')
const { relationshipService }=require('../../services/relationships')
const { normalizeSpaceSelection,publishButtonLabel }=require('../../utils/space-model')

const pad=value=>String(value).padStart(2,'0')
function localParts(value){const date=value?new Date(value):new Date();return{date:`${date.getFullYear()}-${pad(date.getMonth()+1)}-${pad(date.getDate())}`,time:`${pad(date.getHours())}:${pad(date.getMinutes())}`}}
function iso(date,time){return new Date(`${date}T${time}:00`).toISOString()}

// 自动保存防抖：输入 500 字不该产生几百次同步写盘
const SAVE_DEBOUNCE=400

Page({
  data:{text:'',status:'',placeName:'',latitude:null,longitude:null,files:[],occurredAt:'',occurredDate:'',occurredTime:'',today:'',textCount:0,placeCount:0,publishing:false,completed:0,total:0,restored:false,spaceType:'shared',visibility:'couple',publishLabel:'发布到我们的小事',hasPartner:true,statuses:[{code:'in_class',label:'在上课',icon:'✎'},{code:'miss_you',label:'想你了',icon:'♡'},{code:'tired',label:'有点累',icon:'☾'}]},
  onLoad(){
    const draft=draftStore.load()
    const form=restoreDraft(draft)
    const parts=localParts(draft.occurredAt)
    const today=localParts().date
		const space=normalizeSpaceSelection(form)
    this.setData({...form,...space,...parts,occurredDate:parts.date,occurredTime:parts.time,today,textCount:countRunes(form.text),placeCount:countRunes(form.placeName),publishLabel:publishButtonLabel(space)})
    if(!isDraftEmpty(form))this.setData({restored:true})
  },
  async onShow(){if(!sessionStore.snapshot().accessToken)return wx.reLaunch({url:'/pages/welcome/index'});try{const relationship=await relationshipService.getCurrent();this.setData({hasPartner:!relationship.soloTest})}catch(error){if(error.code==='NO_ACTIVE_COUPLE')this.setData({hasPartner:false})}},
  // 离开页面时把没落盘的内容补上；已发布或用户选了「不保存」则不写回
  onUnload(){if(this.suppressPersist)return;this.flushSave()},
  // 拦截左上角返回：用户写了内容还没发布，问他要不要留草稿
  onBackPress(){
    const decision=decideOnExit(this.data,{publishing:this.data.publishing})
    if(decision.action==='leave')return false
    if(decision.action==='blocked'){wx.showToast({title:decision.reason,icon:'none'});return true}
    this.askLeave()
    return true
  },
  askLeave(){
    if(this.asking)return
    this.asking=true
    const hasFiles=this.data.files.length>0
    wx.showModal({
      title:'还没发布哦',
      content:hasFiles?'内容已自动暂存。照片是临时文件，保留草稿也要重新选。':'内容已自动暂存。要保留成草稿，下次接着写吗？',
      confirmText:'保留草稿',
      cancelText:'丢弃',
      confirmColor:'#806AA8',
      success:(result)=>{
        this.asking=false
        if(result.confirm){this.flushSave();if(hasFiles)wx.showToast({title:'文字已保存，照片需重选',icon:'none'})}
        else{this.suppressPersist=true;this.cancelPendingSave();draftStore.clear()}
        wx.navigateBack()
      },
      fail:()=>{this.asking=false;wx.navigateBack()},
    })
  },
  // 用户开始动内容，就不再是「上次那份草稿」了，收起提示条
  markEdited(){if(this.data.restored)this.setData({restored:false})},
  changeSpace(event){const space=normalizeSpaceSelection({spaceType:event.detail.value});this.setData({...space,publishLabel:publishButtonLabel(space)});this.markEdited();this.save()},
  changeVisibility(event){const space=normalizeSpaceSelection({spaceType:'personal',visibility:event.detail.value});this.setData({...space,publishLabel:publishButtonLabel(space)});this.markEdited();this.save()},
  inputText(event){const text=truncateRunes(event.detail.value,500);this.setData({text,textCount:countRunes(text)});this.markEdited();this.save()},
  inputPlace(event){const placeName=truncateRunes(event.detail.value,80);this.setData({placeName,placeCount:countRunes(placeName)});this.markEdited();this.save()},
  selectStatus(event){const code=event.currentTarget.dataset.code;this.setData({status:this.data.status===code?'':code});this.markEdited();this.save()},
  changeDate(event){this.setData({occurredDate:event.detail.value});this.markEdited();this.save()},changeTime(event){this.setData({occurredTime:event.detail.value});this.markEdited();this.save()},
  chooseLocation(){wx.chooseLocation({success:place=>{this.setData({placeName:truncateRunes(place.name||place.address||'',80),placeCount:countRunes(place.name||place.address||''),latitude:place.latitude,longitude:place.longitude});this.markEdited();this.save()}})},
  clearLocation(){this.setData({placeName:'',placeCount:0,latitude:null,longitude:null});this.markEdited();this.save()},
  filesChanged(event){this.queue=null;const files=event.detail.files.map(({path,thumb,size,contentType})=>({path,thumb,size,contentType}));this.setData({files});this.markEdited();this.save()},
  async retryPhoto(event){if(!this.queue||this.data.publishing)return;this.setData({publishing:true});const result=await this.queue.retry(event.detail.index);if(result.some(item=>item.status==='failed')){this.setData({publishing:false});return wx.showToast({title:'还是没传好，再试一次吧',icon:'none'})}await this.completePublish(result)},
  // 自动保存：防抖写入，避免逐字落盘
  save(){if(this.suppressPersist)return;this.cancelPendingSave();this.pendingSave=setTimeout(()=>{this.pendingSave=null;this.persistDraft()},SAVE_DEBOUNCE)},
  cancelPendingSave(){if(this.pendingSave){clearTimeout(this.pendingSave);this.pendingSave=null}},
  persistDraft(){if(this.suppressPersist)return;draftStore.patch(buildDraftPayload({...this.data,occurredAt:iso(this.data.occurredDate,this.data.occurredTime)}))},
  flushSave(){this.cancelPendingSave();if(!this.suppressPersist)this.persistDraft()},
  // 发布成功后把页面清成全新一张，保证下次进来是干净的记录页
  reset(){this.cancelPendingSave();const parts=localParts();this.setData({text:'',status:'',placeName:'',latitude:null,longitude:null,files:[],occurredDate:parts.date,occurredTime:parts.time,textCount:0,placeCount:0,publishing:false,completed:0,total:0,restored:false,spaceType:'shared',visibility:'couple',publishLabel:'发布到我们的小事'})},
  discardDraft(){wx.showModal({title:'清空这份草稿？',content:'上次写的内容会丢掉，无法恢复。',confirmText:'清空',cancelText:'留着',confirmColor:'#C87580',success:(result)=>{if(result.confirm){this.cancelPendingSave();draftStore.clear();this.reset();wx.showToast({title:'草稿已清空',icon:'none'})}}})},
  payload(items){return{text:this.data.text.trim(),status:this.data.status,placeName:this.data.placeName.trim(),latitude:this.data.latitude,longitude:this.data.longitude,occurredAt:iso(this.data.occurredDate,this.data.occurredTime),photoKeys:items.map(item=>item.objectKey),spaceType:this.data.spaceType,visibility:this.data.visibility}},
  async completePublish(items){try{await momentService.publish(this.payload(items));this.suppressPersist=true;this.cancelPendingSave();draftStore.clear();this.queue=null;this.reset();wx.showToast({title:'小事记下了',icon:'success'});setTimeout(()=>wx.navigateBack(),500)}catch(_){this.setData({publishing:false});this.flushSave();wx.showToast({title:'发布失败，草稿已保存',icon:'none'})}},
  async publish(){
    if(this.data.publishing)return
    if(this.data.spaceType==='shared'&&!this.data.hasPartner)return wx.showToast({title:'先绑定伴侣，或发布到我的小事',icon:'none'})
		if(this.data.visibility==='partner'&&!this.data.hasPartner)return wx.showToast({title:'绑定伴侣后才能分享给 TA',icon:'none'})
    if(!this.data.text.trim()&&!this.data.status&&!this.data.placeName.trim()&&!this.data.files.length)return wx.showToast({title:'先记下一点什么吧',icon:'none'})
    this.setData({publishing:true,total:this.data.files.length,completed:0})
    const queue=createUploadQueue({requestGrant:uploadService.requestGrant,putFile:uploadService.putFile,compress:compressSelectedImage,concurrency:2,onChange:items=>this.setData({files:items,completed:items.filter(item=>item.status==='success').length})});this.queue=queue
    try{queue.add(this.data.files);await queue.start();const result=queue.snapshot();if(result.some(item=>item.status==='failed')){this.setData({publishing:false});this.flushSave();return wx.showToast({title:'有照片没传好，点照片重试',icon:'none'})}await this.completePublish(result)}catch(_){this.setData({publishing:false});this.flushSave();wx.showToast({title:'发布失败，草稿已保存',icon:'none'})}
  },
})
