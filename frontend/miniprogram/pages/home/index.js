const { momentService } = require('../../services/moments')
const { sessionStore } = require('../../store/session')
const { storage } = require('../../utils/storage')
const { relationshipService } = require('../../services/relationships')
const { companionshipService } = require('../../services/companionship')
const { statusMoment,normalizeHighlights } = require('../../utils/home-view-model')
const { anniversaryView}=require('../../utils/anniversary')
const { userService } = require('../../services/users')
const { togetherService } = require('../../services/together')
const { reactOptimistically, promptTextReaction, editPerspective, deletePerspective } = require('../../utils/moment-page-actions')
const { memoryResurfaceView } = require('../../utils/memory-resurface')
const { motionModel } = require('../../utils/motion')

const CACHE_KEY = 'homeSnapshot'
// 后端签名图片 URL 有效期 10-15 分钟，缓存必须比它短，
// 否则离线兜底时展示的图片会因签名过期返回 403。
const CACHE_TTL_MS = 5 * 60 * 1000

function readFreshCache() {
  const entry = storage.get(CACHE_KEY)
  if (!entry || typeof entry !== 'object' || !entry.savedAt) return null
  if (Date.now() - Number(entry.savedAt) > CACHE_TTL_MS) {
    storage.remove(CACHE_KEY)
    return null
  }
  return entry.payload
}

Page({
  data:{state:'loading',home:null,moments:[],pinnedAnniversaries:[],hasMorePinned:false,dailyQuestion:null,resurfacedMemory:null,inboxExpanded:false,relationship:null,currentUser:{},companionship:null,presenceMoment:{},sendingNudge:'',partnerInitial:'♡',relationshipDays:0,currentUserId:0,currentUserAvatar:'',partnerAvatar:'',motionClass:'',error:''},
  onLoad(){ const session=sessionStore.snapshot();const user=session.user||{};this.setData({currentUser:user,currentUserId:Number(session.userId)||0,currentUserAvatar:user.avatarUrl||'',motionClass:motionModel(getApp().globalData.reduceMotion).className}) },
  async onShow(){ const session=sessionStore.snapshot();if(!session.accessToken)return wx.reLaunch({url:'/pages/welcome/index'});const user=session.user||{};this.setData({currentUser:user,currentUserAvatar:user.avatarUrl||''});let relationship;try{relationship=await relationshipService.getCurrent()}catch(error){if(error.code==='NO_ACTIVE_COUPLE'){storage.remove(CACHE_KEY);this.setData({home:null,moments:[],pinnedAnniversaries:[]});return wx.reLaunch({url:'/pages/bind/index'})}return}const partner=relationship.partner||{};const boundAt=new Date(relationship.boundAt);const days=Number.isNaN(boundAt.getTime())?0:Math.max(1,Math.floor((Date.now()-boundAt.getTime())/86400000)+1);this.setData({relationship,partnerInitial:(partner.displayName||'♡').slice(0,1),partnerAvatar:partner.avatarUrl||'',relationshipDays:days});const cached=readFreshCache();if(!this.data.home&&cached)this.setData({home:cached,moments:cached.recent||[],pinnedAnniversaries:normalizeHighlights(cached.pinnedAnniversaries).map(item=>anniversaryView(item)),hasMorePinned:Boolean(cached.hasMorePinned)});this.refreshAvatar();this.load({quiet:Boolean(cached)}) },
  onPullDownRefresh(){ this.refreshAvatar(); this.load().finally(()=>wx.stopPullDownRefresh()) },
  // 头像 URL 是后端签名的（约 15 分钟过期），本地session 里存的是旧签名，
  // 必须重新拉取，否则超过有效期后头像会一直加载失败。
  async refreshAvatar(){
    try{
      const fresh=await userService.getMe()
      const session=sessionStore.setUser(fresh)
      this.setData({currentUser:fresh,currentUserAvatar:fresh.avatarUrl||''})
      return session
    }catch(error){ return null }
  },
  async load({quiet=false}={}){
    if(!quiet&&!this.data.home) this.setData({state:'loading',error:''})
    try{ const [home,companionship,dailyQuestion,memory]=await Promise.all([momentService.getHome(),companionshipService.getSnapshot(),togetherService.getDailyQuestion().catch(()=>null),momentService.getResurfacedMemory().catch(()=>null)]); const moments=home.recent||[];const pinnedAnniversaries=normalizeHighlights(home.pinnedAnniversaries).map(item=>anniversaryView(item));const presenceMoment=statusMoment(companionship.partnerPresence,home.partnerLatestStatus); storage.set(CACHE_KEY,{savedAt:Date.now(),payload:home}); this.setData({home,companionship,dailyQuestion,resurfacedMemory:memoryResurfaceView(memory),presenceMoment,moments,pinnedAnniversaries,hasMorePinned:Boolean(home.hasMorePinned),state:moments.length?'ready':'empty',error:''}) }
    catch(error){ const cached=this.data.home; this.setData({state:cached?'error-with-cache':'error-empty',error:'暂时没有连上，再试一次吧'}) }
  },
  retry(){this.load()},
  editStatus(){wx.navigateTo({url:'/pages/status-editor/index'})},
  async sendNudge(event){const kind=event.currentTarget.dataset.kind;if(this.data.sendingNudge)return;this.setData({sendingNudge:kind});try{await companionshipService.sendNudge(kind);wx.showToast({title:kind==='hug'?'抱抱送到啦':kind==='cheer_up'?'给 TA 加油啦':'想念送到啦',icon:'none'})}catch(error){wx.showToast({title:error.code==='NUDGE_THROTTLED'?'慢一点，让心意先飞一会儿':'发送失败，稍后再试',icon:'none'})}finally{this.setData({sendingNudge:''})}},
  compose(){wx.navigateTo({url:'/pages/compose/index'})},
  openMomentSpace(scope){storage.set('momentSpaceScope',scope);wx.switchTab({url:'/pages/moment-spaces/index'})},
  openSharedMoments(){this.openMomentSpace('shared')},
  openMyMoments(){this.openMomentSpace('mine')},
  openPartnerMoments(){if(this.data.relationship&&this.data.relationship.soloTest)return wx.showToast({title:'单人测试模式没有 TA 空间',icon:'none'});this.openMomentSpace('partner')},
  toggleInbox(){this.setData({inboxExpanded:!this.data.inboxExpanded})},
  openAnniversary(event){wx.navigateTo({url:`/pages/anniversary-detail/index?id=${event.detail.id}`})},
  openAnniversaries(){wx.navigateTo({url:'/pages/anniversaries/index'})},
  preview(event){wx.previewImage(event.detail)},
  edit(event){storage.set('editingMoment',event.detail.moment);wx.navigateTo({url:'/pages/moment-edit/index'})},
  remove(event){
    const id=event.detail.id
    wx.showModal({title:'删除这件小事？',content:'删除后不能恢复。',confirmText:'删除',confirmColor:'#C87580',success:async(result)=>{if(!result.confirm)return;try{await momentService.deleteMoment(id);this.setData({moments:this.data.moments.filter((item)=>item.id!==id)});wx.showToast({title:'已删除',icon:'none'})}catch(_){wx.showToast({title:'删除失败，请稍后再试',icon:'none'})}}})
  },
  replaceMoment(moment){this.setData({moments:this.data.moments.map(item=>String(item.id)===String(moment.id)?moment:item)})},
  react(event){reactOptimistically(event,this.data.currentUserId,moment=>this.replaceMoment(moment))},
  textReaction(event){promptTextReaction(event,moment=>this.replaceMoment(moment))},
  editPerspective(event){editPerspective(event)},
  deletePerspective(event){deletePerspective(event,moment=>this.replaceMoment(moment))},
  async submitDailyAnswer(event){try{const dailyQuestion=await togetherService.submitDailyAnswer({questionId:this.data.dailyQuestion.id,text:event.detail.text});this.setData({dailyQuestion});event.detail.done();if(dailyQuestion.unlocked)this.load({quiet:true})}catch(error){event.detail.done();wx.showToast({title:'答案没有藏好，再试一次吧',icon:'none'})}},
  openDailyQuestions(){wx.navigateTo({url:'/pages/daily-questions/index'})},
  openResurfacedMemory(){wx.switchTab({url:'/pages/timeline/index'})},
})
