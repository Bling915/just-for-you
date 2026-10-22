const {companionshipService}=require('../../services/companionship')
const {countRunes,truncateRunes}=require('../../utils/runes')
const {messageForError}=require('../../utils/error-copy')
Page({
 data:{presets:[{code:'in_class',label:'在上课',icon:'✎'},{code:'miss_you',label:'想你了',icon:'♡'},{code:'tired',label:'有点累',icon:'☾'},{code:'custom',label:'自定义',icon:'✦'}],durations:[{minutes:60,label:'1 小时'},{minutes:120,label:'2 小时'},{minutes:240,label:'4 小时'},{minutes:0,label:'一直'}],preset:'in_class',customText:'',duration:60,submitting:false,error:''},
 choose(event){this.setData({preset:event.currentTarget.dataset.code,error:''})},
 inputCustom(event){this.setData({customText:truncateRunes(event.detail.value,20),error:''})},
 chooseDuration(event){this.setData({duration:Number(event.currentTarget.dataset.minutes)||0})},
 async submit(){if(this.data.submitting)return;if(this.data.preset==='custom'&&!countRunes(this.data.customText.trim()))return this.setData({error:'写一句此刻的状态吧'});const expiresAt=this.data.duration?new Date(Date.now()+this.data.duration*60000).toISOString():null;this.setData({submitting:true,error:''});try{await companionshipService.updatePresence({preset:this.data.preset,customText:this.data.preset==='custom'?this.data.customText:'',expiresAt});wx.showToast({title:'状态更新啦',icon:'none'});setTimeout(()=>wx.navigateBack(),350)}catch(error){this.setData({error:messageForError(error.code)})}finally{this.setData({submitting:false})}}
})
