package domain

import "errors"

var (
	ErrInvalidNickname     = errors.New("invalid nickname")
	ErrCodeRequired        = errors.New("wechat code is required")
	ErrInvalidRefreshToken = errors.New("refresh token is invalid")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidAvatar       = errors.New("invalid avatar")
)
