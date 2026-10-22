const { http } = require('./http-client');

const authService = {
  loginWithWeChat: (code) => http.post('/v1/auth/wechat', { code }),
  refresh: (refreshToken) => http.post('/v1/auth/refresh', { refreshToken }),
  logout: (refreshToken) => http.post('/v1/auth/logout', { refreshToken }),
};

module.exports = { authService };
