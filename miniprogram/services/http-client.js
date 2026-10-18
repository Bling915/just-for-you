const { APP_CONFIG } = require('../config.js');
const { sessionStore } = require('../store/session');

function toError(response) {
  const body = response && response.data && response.data.error;
  return {
    status: (response && response.statusCode) || 0,
    code: (body && body.code) || 'NETWORK_ERROR',
    message: (body && body.message) || '网络连接失败，请稍后重试',
  };
}

function createHttpClient({ request, storage, onSessionExpired, baseURL }) {
  let refreshPromise = null;

  async function raw(method, path, data, accessToken) {
    const header = { 'Content-Type': 'application/json' };
    if (accessToken) header.Authorization = `Bearer ${accessToken}`;
    try {
      return await request({ url: `${baseURL}${path}`, method, data, header });
    } catch (_) {
      throw { status: 0, code: 'NETWORK_ERROR', message: '网络连接失败，请稍后重试' };
    }
  }

  function refreshAccess() {
    if (refreshPromise) return refreshPromise;
    const refreshState = storage.snapshot();
    const refreshEpoch = refreshState.epoch;
    refreshPromise = (async () => {
      const refreshToken = refreshState.refreshToken;
      if (!refreshToken) throw { status: 401, code: 'INVALID_REFRESH', message: '请重新登录' };
      const response = await raw('POST', '/v1/auth/refresh', { refreshToken }, '');
      if (response.statusCode < 200 || response.statusCode >= 300) throw toError(response);
      const session = response.data && response.data.data;
      if (storage.setSession(session, refreshEpoch) === false) throw { status: 401, code: 'SESSION_CHANGED', message: '会话已变化' };
      return session.accessToken;
    })().catch((error) => {
      if (storage.snapshot().epoch === refreshEpoch) {
        storage.clear();
        onSessionExpired();
      }
      throw error;
    }).finally(() => { refreshPromise = null; });
    return refreshPromise;
  }

  async function execute(method, path, data, mayRefresh = true) {
    const response = await raw(method, path, data, storage.snapshot().accessToken);
    if (response.statusCode === 401 && mayRefresh && path !== '/v1/auth/refresh') {
      await refreshAccess();
      return execute(method, path, data, false);
    }
    if (response.statusCode < 200 || response.statusCode >= 300) throw toError(response);
    return response.data ? response.data.data : undefined;
  }

  return {
    get: (path) => execute('GET', path),
    post: (path, data) => execute('POST', path, data),
    put: (path, data) => execute('PUT', path, data),
    delete: (path, data) => execute('DELETE', path, data),
  };
}

function wxRequest(options) {
  return new Promise((resolve, reject) => wx.request({ ...options, success: resolve, fail: reject }));
}

const http = createHttpClient({
  request: wxRequest,
  storage: sessionStore,
  onSessionExpired() {
    const { appState } = require('../store/app-state');
    appState.clearPrivateData();
    wx.reLaunch({ url: '/pages/welcome/index' });
  },
  baseURL: APP_CONFIG.apiBaseURL,
});

module.exports = { createHttpClient, http };
