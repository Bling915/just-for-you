package domain

import "time"

type RefreshToken struct {
	hash      [32]byte
	userID    int64
	expiresAt time.Time
	revokedAt *time.Time
}

func NewRefreshToken(hash [32]byte, userID int64, expiresAt time.Time) RefreshToken {
	return RefreshToken{hash: hash, userID: userID, expiresAt: expiresAt}
}

func RestoreRefreshToken(hash [32]byte, userID int64, expiresAt time.Time, revokedAt *time.Time) RefreshToken {
	return RefreshToken{hash: hash, userID: userID, expiresAt: expiresAt, revokedAt: revokedAt}
}

func (t RefreshToken) Hash() [32]byte        { return t.hash }
func (t RefreshToken) UserID() int64         { return t.userID }
func (t RefreshToken) ExpiresAt() time.Time  { return t.expiresAt }
func (t RefreshToken) RevokedAt() *time.Time { return t.revokedAt }
func (t RefreshToken) IsUsableAt(now time.Time) bool {
	return t.revokedAt == nil && t.expiresAt.After(now)
}
func (t *RefreshToken) Revoke(now time.Time) { t.revokedAt = &now }
