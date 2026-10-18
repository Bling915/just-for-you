const { http } = require('./http-client');

const userService = {
  getMe: () => http.get('/v1/users/me'),
  updateProfile: (input) => http.put('/v1/users/me', input),
};

module.exports = { userService };
