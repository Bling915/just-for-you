const { momentService } = require('../../services/moments')
const { sessionStore } = require('../../store/session')
const { storage } = require('../../utils/storage')
const { relationshipService } = require('../../services/relationships')

const CACHE_KEY = 'homeSnapshot'

Page({
  data:{state:'loading',home:null,moments:[],relationship:null,partnerInitial:'♡',relationshipDays:0,currentUserId:0,error:''},
  onLoad(){ const session=sessionStore.snapshot(); this.setData({currentUserId:Number(session.userId)||0}) },
  async onShow(){ const session=sessionStore.snapshot();if(!session.accessToken)return wx.reLaunch({url:'/pages/welcome/index'});let relationship;try{relationship=await relationshipService.getCurrent()}catch(error){if(error.code==='NO_ACTIVE_COUPLE'){storage.remove(CACHE_KEY);this.setData({home:null,moments:[]});return wx.reLaunch({url:'/pages/bind/index'})}return}const partner=relationship.partner||{};const boundAt=new Date(relationship.boundAt);const days=Number.isNaN(boundAt.getTime())?0:Math.max(1,Math.floor((Date.now()-boundAt.getTime())/86400000)+1);this.setData({relationship,partnerInitial:(partner.displayName||'♡').slice(0,1),relationshipDays:days});const cached=storage.get(CACHE_KEY);if(!this.data.home&&cached)this.setData({home:cached,moments:cached.recent||[]});this.load({quiet:Boolean(cached)}) },
  onPullDownRefresh(){ this.load().finally(()=>wx.stopPullDownRefresh()) },
  async load({quiet=false}={}){
    if(!quiet&&!this.data.home) this.setData({state:'loading',error:''})
    try{ const home=await momentService.getHome(); const moments=home.recent||[]; storage.set(CACHE_KEY,home); this.setData({home,moments,state:moments.length?'ready':'empty',error:''}) }
    catch(error){ const cached=this.data.home; this.setData({state:cached?'error-with-cache':'error-empty',error:'暂时没有连上，再试一次吧'}) }
  },
  retry(){this.load()},
  compose(){wx.navigateTo({url:'/pages/compose/index'})},
  preview(event){wx.previewImage(event.detail)},
  remove(event){
    const id=event.detail.id
    wx.showModal({title:'删除这件小事？',content:'删除后不能恢复。',confirmText:'删除',confirmColor:'#C87580',success:async(result)=>{if(!result.confirm)return;try{await momentService.deleteMoment(id);this.setData({moments:this.data.moments.filter((item)=>item.id!==id)});wx.showToast({title:'已删除',icon:'none'})}catch(_){wx.showToast({title:'删除失败，请稍后再试',icon:'none'})}}})
  },
})
