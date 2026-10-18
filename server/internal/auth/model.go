package auth

import (
	"errors"
	"time"
)

var ErrCodeRequired = errors.New("wechat code is required")
var ErrInvalidRefresh = errors.New("refresh token is invalid")

type Session struct {
	UserID          int64     `json:"userId"`
	AccessToken     string    `json:"accessToken"`
	RefreshToken    string    `json:"refreshToken"`
	AccessExpiresAt time.Time `json:"accessExpiresAt"`
}
type RefreshRecord struct {
	Hash      [32]byte
	UserID    int64
	ExpiresAt time.Time
	RevokedAt *time.Time
}
