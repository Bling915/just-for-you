const { createTimelineStore } = require('../../store/timeline')
const { momentService } = require('../../services/moments')
const { sessionStore } = require('../../store/session')
const { relationshipService } = require('../../services/relationships')
const { storage } = require('../../utils/storage')
const { reactOptimistically, promptTextReaction, shareToCouple } = require('../../utils/moment-page-actions')
const { motionModel } = require('../../utils/motion')

const store = createTimelineStore(momentService, { space: 'personal', owner: 'me' })

Page({
  data: { owner: 'me', title: '我的小事', subtitle: '只属于你的日常角落', moments: [], loading: true, loadingMore: false, hasMore: true, error: '', currentUserId: 0, currentUserAvatar: '', partnerAvatar: '',motionClass:'' },
  onLoad(options) {
    const owner = options.owner === 'partner' ? 'partner' : 'me'
    const session = sessionStore.snapshot()
    store.reset({ space: 'personal', owner })
    this.setData({ owner, title: owner === 'partner' ? 'TA 的小事' : '我的小事', subtitle: owner === 'partner' ? 'TA 愿意与你分享的片刻' : '安静收藏，也可以分享给 TA', currentUserId: Number(session.userId) || 0, currentUserAvatar: (session.user && session.user.avatarUrl) || '',motionClass:motionModel(getApp().globalData.reduceMotion).className })
  },
  async onShow() {
    if (!sessionStore.snapshot().accessToken) return wx.reLaunch({ url: '/pages/welcome/index' })
    try {
      const relationship = await relationshipService.getCurrent()
      this.setData({ partnerAvatar: (relationship.partner && relationship.partner.avatarUrl) || '' })
    } catch (error) {
      if (this.data.owner === 'partner') {
        this.setData({ loading: false, error: '绑定伴侣后才能查看 TA 的小事' })
        return
      }
    }
    await this.refresh()
  },
  onPullDownRefresh() { this.refresh().finally(() => wx.stopPullDownRefresh()) },
  onReachBottom() { if (!this.data.loading && !this.data.loadingMore && this.data.hasMore) this.more() },
  sync() { this.setData(store.snapshot()) },
  async refresh() { store.reset({ space: 'personal', owner: this.data.owner }); this.setData({ loading: true, error: '' }); try { await store.loadFirst() } catch (_) {} this.sync(); this.setData({ loading: false }) },
  async more() { this.setData({ loadingMore: true }); try { await store.loadMore() } catch (_) { wx.showToast({ title: '加载失败，稍后再试', icon: 'none' }) } this.sync(); this.setData({ loadingMore: false }) },
  compose() { wx.navigateTo({ url: '/pages/compose/index' }) },
  preview(event) { wx.previewImage(event.detail) },
  edit(event) { storage.set('editingMoment', event.detail.moment); wx.navigateTo({ url: '/pages/moment-edit/index' }) },
  remove(event) {
    const id = event.detail.id
    wx.showModal({ title: '删除这件小事？', content: '删除后不能恢复。', confirmText: '删除', confirmColor: '#C87580', success: async result => {
      if (!result.confirm) return
      try { await momentService.deleteMoment(id); store.remove(id); this.sync(); wx.showToast({ title: '已删除', icon: 'none' }) }
      catch (_) { wx.showToast({ title: '删除失败，请稍后再试', icon: 'none' }) }
    } })
  },
  react(event) { reactOptimistically(event, this.data.currentUserId, moment => { store.replace(moment); this.sync() }) },
  textReaction(event) { promptTextReaction(event, moment => { store.replace(moment); this.sync() }) },
  share(event) { shareToCouple(event) },
})
