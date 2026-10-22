const { compressSelectedImage } = require('./image-compression');
const { putFile } = require('../services/uploads');
const { userService } = require('../services/users');

function getSize(filePath) {
  return new Promise((resolve, reject) => {
    wx.getFileInfo({
      filePath,
      success: (result) => resolve(result.size),
      fail: reject,
    });
  });
}

function contentTypeFor(filePath) {
  if (/\.png$/i.test(filePath)) return 'image/png';
  if (/\.webp$/i.test(filePath)) return 'image/webp';
  return 'image/jpeg';
}

function createAvatarReplacer(deps) {
  return async function replaceAvatar({ filePath, onProgress = () => {} }) {
    if (!filePath) throw Object.assign(new Error('请选择头像'), { code: 'EMPTY_AVATAR' });

    const compressed = await deps.compress(filePath);
    const size = await deps.getSize(compressed);
    const profile = await deps.getMe();
    const contentType = contentTypeFor(compressed);
    const grant = await deps.requestGrant({ contentType, size });
    await deps.putFile({
      filePath: compressed,
      uploadURL: grant.uploadURL,
      contentType,
      onProgress,
    });
    return deps.updateProfile({ displayName: profile.displayName, avatarRef: grant.objectKey });
  };
}

const replaceAvatar = createAvatarReplacer({
  compress: compressSelectedImage,
  getSize,
  getMe: userService.getMe,
  requestGrant: userService.requestAvatarUpload,
  putFile,
  updateProfile: userService.updateProfile,
});

module.exports = { createAvatarReplacer, replaceAvatar, contentTypeFor };
