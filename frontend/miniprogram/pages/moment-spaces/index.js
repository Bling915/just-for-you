const { createTimelineStore } = require('../../store/timeline')
const { momentService } = require('../../services/moments')
const { sessionStore } = require('../../store/session')
const { relationshipService } = require('../../services/relationships')
const { storage } = require('../../utils/storage')
const { reactOptimistically, promptTextReaction, shareToCouple, editPerspective, deletePerspective } = require('../../utils/moment-page-actions')

const store = createTimelineStore(momentService)
const scopes = {
  shared: { space: 'shared' },
  mine: { space: 'personal', owner: 'me' },
  partner: { space: 'personal', owner: 'partner' },
}

Page({
  data: { scope: 'shared', moments: [], loading: true, loadingMore: false, hasMore: true, error: '', currentUserId: 0, currentUserAvatar: '', partnerAvatar: '', soloTest: false },
  onLoad(options) {
    const scope = scopes[options.scope] ? options.scope : 'shared'
    const session = sessionStore.snapshot()
    this.setData({ scope, currentUserId: Number(session.userId) || 0, currentUserAvatar: (session.user && session.user.avatarUrl) || '' })
    store.reset(scopes[scope])
  },
  async onShow() {
    if (!sessionStore.snapshot().accessToken) return wx.reLaunch({ url: '/pages/welcome/index' })
    try {
      const relationship = await relationshipService.getCurrent()
      this.setData({ partnerAvatar: (relationship.partner && relationship.partner.avatarUrl) || '', soloTest: Boolean(relationship.soloTest) })
      if (relationship.soloTest && this.data.scope === 'partner') this.setData({ scope: 'mine' })
    } catch (_) {}
    await this.refresh()
  },
  changeScope(event) {
    const scope = event.currentTarget.dataset.scope
    if (!scopes[scope] || scope === this.data.scope) return
    if (scope === 'partner' && this.data.soloTest) return wx.showToast({ title: '单人测试模式没有 TA 空间', icon: 'none' })
    this.setData({ scope })
    this.refresh()
  },
  onPullDownRefresh() { this.refresh().finally(() => wx.stopPullDownRefresh()) },
  onReachBottom() { if (!this.data.loading && !this.data.loadingMore && this.data.hasMore) this.more() },
  sync() { this.setData(store.snapshot()) },
  async refresh() { store.reset(scopes[this.data.scope]); this.setData({ loading: true, error: '' }); try { await store.loadFirst() } catch (_) {} this.sync(); this.setData({ loading: false }) },
  async more() { this.setData({ loadingMore: true }); try { await store.loadMore() } catch (_) { wx.showToast({ title: '加载失败，稍后再试', icon: 'none' }) } this.sync(); this.setData({ loadingMore: false }) },
  compose() { wx.navigateTo({ url: '/pages/compose/index' }) },
  preview(event) { wx.previewImage(event.detail) },
  edit(event) { storage.set('editingMoment', event.detail.moment); wx.navigateTo({ url: '/pages/moment-edit/index' }) },
  remove(event) { const id = event.detail.id; wx.showModal({ title: '删除这件小事？', content: '删除后不能恢复。', confirmText: '删除', confirmColor: '#C87580', success: async result => { if (!result.confirm) return; try { await momentService.deleteMoment(id); store.remove(id); this.sync(); wx.showToast({ title: '已删除', icon: 'none' }) } catch (_) { wx.showToast({ title: '删除失败，请稍后再试', icon: 'none' }) } } }) },
  react(event) { reactOptimistically(event, this.data.currentUserId, moment => { store.replace(moment); this.sync() }) },
  textReaction(event) { promptTextReaction(event, moment => { store.replace(moment); this.sync() }) },
  share(event) { shareToCouple(event, () => { if (this.data.scope === 'mine') this.refresh() }) },
  editPerspective(event) { editPerspective(event) },
  deletePerspective(event) { deletePerspective(event, moment => { store.replace(moment); this.sync() }) },
})
