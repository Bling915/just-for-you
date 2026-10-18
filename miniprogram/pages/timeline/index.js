const { timelineStore }=require('../../store/timeline');const { momentService }=require('../../services/moments');const { sessionStore }=require('../../store/session');const { relationshipService }=require('../../services/relationships')
Page({
 data:{moments:[],loading:true,loadingMore:false,hasMore:true,error:'',currentUserId:0},
 onLoad(){this.setData({currentUserId:Number(sessionStore.snapshot().userId)||0})},
 async onShow(){const session=sessionStore.snapshot();if(!session.accessToken)return wx.reLaunch({url:'/pages/welcome/index'});try{await relationshipService.getCurrent()}catch(error){timelineStore.reset();this.sync();if(error.code==='NO_ACTIVE_COUPLE')return wx.reLaunch({url:'/pages/bind/index'});return}if(!this.initialized){this.initialized=true;this.refresh()}},
 onPullDownRefresh(){this.refresh().finally(()=>wx.stopPullDownRefresh())},
 onReachBottom(){if(!this.data.loading&&!this.data.loadingMore&&this.data.hasMore)this.more()},
 sync(){this.setData(timelineStore.snapshot())},
 async refresh(){timelineStore.reset();this.setData({loading:true,error:''});try{await timelineStore.loadFirst()}catch(_){}this.sync();this.setData({loading:false})},
 async more(){this.setData({loadingMore:true});try{await timelineStore.loadMore()}catch(_){wx.showToast({title:'加载失败，稍后再试',icon:'none'})}this.sync();this.setData({loadingMore:false})},
 retry(){this.refresh()},compose(){wx.navigateTo({url:'/pages/compose/index'})},preview(event){wx.previewImage(event.detail)},
 remove(event){const id=event.detail.id;wx.showModal({title:'删除这件小事？',content:'删除后不能恢复。',confirmText:'删除',confirmColor:'#C87580',success:async result=>{if(!result.confirm)return;try{await momentService.deleteMoment(id);timelineStore.remove(id);this.sync();wx.showToast({title:'已删除',icon:'none'})}catch(error){if(error.code==='NOT_AUTHOR'||error.code==='MOMENT_NOT_FOUND'){await this.refresh();wx.showToast({title:error.code==='NOT_AUTHOR'?'只能删除自己的记录':'这条记录已经不存在了',icon:'none'})}else wx.showToast({title:'删除失败，请稍后再试',icon:'none'})}}})},
})
