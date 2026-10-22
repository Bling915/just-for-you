const { momentService } = require('../services/moments')
const { mergeMomentPage } = require('../utils/merge-page')

function normalizeFilter(filter = {}) {
  return filter.space === 'personal'
    ? { space: 'personal', owner: filter.owner === 'partner' ? 'partner' : 'me' }
    : { space: 'shared', owner: '' }
}

function createTimelineStore(service, initialFilter = { space: 'shared' }) {
  let state = { moments: [], cursor: '', hasMore: true, loading: false, error: '', filter: normalizeFilter(initialFilter) }
  let inflight = null
  let generation = 0
  const snapshot = () => ({ ...state, filter: { ...state.filter }, moments: state.moments.map(item => ({ ...item })) })

  function reset(nextFilter) {
    generation += 1
    inflight = null
    state = {
      moments: [], cursor: '', hasMore: true, loading: false, error: '',
      filter: normalizeFilter(nextFilter || state.filter),
    }
  }

  async function fetchPage(first) {
    if (inflight) return inflight
    if (!first && !state.hasMore) return snapshot()
    const startGeneration = generation
    const requestFilter = { ...state.filter }
    state = { ...state, loading: true, error: '' }
    inflight = (async () => {
      try {
        const page = await service.getTimeline({ ...requestFilter, cursor: first ? '' : state.cursor, limit: 20 })
        if (startGeneration !== generation) return snapshot()
        state = {
          ...state,
          moments: first ? (page.moments || []) : mergeMomentPage(state.moments, page.moments || []),
          cursor: page.nextCursor || '',
          hasMore: Boolean(page.nextCursor),
          loading: false,
          error: '',
        }
        return snapshot()
      } catch (error) {
        if (startGeneration === generation) state = { ...state, loading: false, error: (error && error.code) || 'LOAD_FAILED' }
        throw error
      } finally {
        if (startGeneration === generation) inflight = null
      }
    })()
    return inflight
  }

  return {
    reset,
    setFilter(filter) { reset(filter); return snapshot() },
    loadFirst() { return fetchPage(true) },
    loadMore() { return fetchPage(false) },
    remove(id) { state = { ...state, moments: state.moments.filter(item => String(item.id) !== String(id)) } },
    replace(moment) { state = { ...state, moments: state.moments.map(item => String(item.id) === String(moment.id) ? moment : item) } },
    snapshot,
  }
}

const timelineStore = createTimelineStore(momentService, { space: 'shared' })
module.exports = { createTimelineStore, timelineStore }
