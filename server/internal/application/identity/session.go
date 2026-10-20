package identity

import (
	"context"
	"crypto/sha256"
	"time"

	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
)

type Session struct {
	UserID          int64     `json:"userId"`
	AccessToken     string    `json:"accessToken"`
	RefreshToken    string    `json:"refreshToken"`
	AccessExpiresAt time.Time `json:"accessExpiresAt"`
}

type sessionIssuer struct {
	repository      Repository
	tokens          AccessTokenIssuer
	clock           appshared.Clock
	generateRefresh func() (string, error)
}

func (s sessionIssuer) create(ctx context.Context, userID int64) (Session, error) {
	return s.issue(ctx, userID, nil)
}

func (s sessionIssuer) rotate(ctx context.Context, raw string) (Session, error) {
	hash := sha256.Sum256([]byte(raw))
	return s.issue(ctx, 0, &hash)
}

func (s sessionIssuer) issue(ctx context.Context, userID int64, old *[32]byte) (Session, error) {
	now := s.clock.Now()
	raw, err := s.generateRefresh()
	if err != nil {
		return Session{}, err
	}
	next := domain.NewRefreshToken(sha256.Sum256([]byte(raw)), userID, now.Add(30*24*time.Hour))
	if old == nil {
		err = s.repository.StoreRefresh(ctx, next)
	} else {
		userID, err = s.repository.RotateRefresh(ctx, *old, next, now)
	}
	if err != nil {
		return Session{}, err
	}
	access, expiresAt, err := s.tokens.IssueAccess(userID, now)
	if err != nil {
		return Session{}, err
	}
	return Session{UserID: userID, AccessToken: access, RefreshToken: raw, AccessExpiresAt: expiresAt}, nil
}
