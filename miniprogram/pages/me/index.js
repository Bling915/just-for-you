const { sessionStore }=require('../../store/session');const { appState }=require('../../store/app-state');const { relationshipService }=require('../../services/relationships');const { authService }=require('../../services/auth')
Page({
 data:{user:{},initial:'♡',relationship:null,boundDate:'',relationshipDays:0,motionClass:'',loading:true},
 async onShow(){const session=sessionStore.snapshot();if(!session.accessToken)return wx.reLaunch({url:'/pages/welcome/index'});const user=session.user||{};this.setData({user,initial:(user.displayName||'♡').slice(0,1),motionClass:getApp().globalData.reduceMotion?'reduce-motion':'',loading:true});try{const relationship=await relationshipService.getCurrent();const boundAt=new Date(relationship.boundAt);const relationshipDays=Number.isNaN(boundAt.getTime())?0:Math.max(1,Math.floor((Date.now()-boundAt.getTime())/86400000)+1);this.setData({relationship,relationshipDays,boundDate:(relationship.boundAt||'').slice(0,10),loading:false})}catch(error){this.setData({relationship:null,boundDate:'',relationshipDays:0,loading:false});if(error.code==='NO_ACTIVE_COUPLE')wx.reLaunch({url:'/pages/bind/index'})}},
 edit(){wx.navigateTo({url:'/pages/profile-setup/index?edit=1'})},
 avatarChanged(event){const user=event.detail.user;this.setData({user,initial:(user.displayName||'♡').slice(0,1)})},
 openTogether(event){wx.navigateTo({url:event.currentTarget.dataset.url})},
 openNotifications(){wx.navigateTo({url:'/pages/notification-settings/index'})},
 endRelationship(){wx.showModal({title:'解除情侣关系？',content:'解除后，共同时间线和照片会立即不可见，且无法从小程序恢复。',confirmText:'确认解除',confirmColor:'#C87580',success:async result=>{if(!result.confirm)return;try{await relationshipService.end();appState.clearRelationshipData();wx.reLaunch({url:'/pages/bind/index'})}catch(_){wx.showToast({title:'解除失败，请稍后再试',icon:'none'})}}})},
 logout(){wx.showModal({title:'退出登录？',content:'本机保存的账号信息和草稿会被清除。',confirmText:'退出',confirmColor:'#C87580',success:async result=>{if(!result.confirm)return;const refreshToken=sessionStore.snapshot().refreshToken;try{await authService.logout(refreshToken)}catch(_){}appState.clearPrivateData();wx.reLaunch({url:'/pages/welcome/index'})}})},
})
