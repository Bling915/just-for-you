function createWxStorage(wxAPI) {
  return {
    get(key) {
      try { return wxAPI.getStorageSync(key); } catch (_) { return undefined; }
    },
    set(key, value) { wxAPI.setStorageSync(key, value); },
    remove(key) { wxAPI.removeStorageSync(key); },
  };
}

const storage = createWxStorage({
  getStorageSync: (...args) => wx.getStorageSync(...args),
  setStorageSync: (...args) => wx.setStorageSync(...args),
  removeStorageSync: (...args) => wx.removeStorageSync(...args),
});

module.exports = { createWxStorage, storage };
