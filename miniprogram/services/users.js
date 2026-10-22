const { http } = require('./http-client');

const userService = {
  getMe: () => http.get('/v1/users/me'),
  updateProfile: (input) => http.put('/v1/users/me', input),
  requestAvatarUpload: (input) => http.post('/v1/users/me/avatar-upload', input),
};

module.exports = { userService };
