package application

import (
	"context"
	"crypto/sha256"
	"strings"
	"time"

	"github.com/nanfang/just-for-you/server/internal/identity/domain"
	platformclock "github.com/nanfang/just-for-you/server/internal/platform/clock"
)

type Session struct {
	UserID          int64     `json:"userId"`
	AccessToken     string    `json:"accessToken"`
	RefreshToken    string    `json:"refreshToken"`
	AccessExpiresAt time.Time `json:"accessExpiresAt"`
}

type Service struct {
	provider        IdentityProvider
	repository      Repository
	tokens          AccessTokenIssuer
	clock           platformclock.Clock
	generateRefresh func() (string, error)
}

func NewService(provider IdentityProvider, repository Repository, tokens AccessTokenIssuer, clock platformclock.Clock, generator func() (string, error)) *Service {
	return &Service{provider: provider, repository: repository, tokens: tokens, clock: clock, generateRefresh: generator}
}

func (s *Service) LoginWithWeChat(ctx context.Context, code string) (Session, error) {
	if strings.TrimSpace(code) == "" {
		return Session{}, domain.ErrCodeRequired
	}
	external, err := s.provider.Exchange(ctx, code)
	if err != nil {
		return Session{}, err
	}
	user, err := s.repository.FindOrCreateUser(ctx, external.OpenID)
	if err != nil {
		return Session{}, err
	}
	return s.newSession(ctx, user.ID(), nil)
}

func (s *Service) newSession(ctx context.Context, userID int64, old *[32]byte) (Session, error) {
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
	access, expires, err := s.tokens.IssueAccess(userID, now)
	if err != nil {
		return Session{}, err
	}
	return Session{UserID: userID, AccessToken: access, RefreshToken: raw, AccessExpiresAt: expires}, nil
}

func (s *Service) RefreshSession(ctx context.Context, raw string) (Session, error) {
	hash := sha256.Sum256([]byte(raw))
	return s.newSession(ctx, 0, &hash)
}
func (s *Service) Logout(ctx context.Context, raw string) error {
	hash := sha256.Sum256([]byte(raw))
	return s.repository.RevokeRefresh(ctx, hash, s.clock.Now())
}
func (s *Service) GetMe(ctx context.Context, userID int64) (domain.User, error) {
	return s.repository.GetUser(ctx, userID)
}
func (s *Service) UpdateProfile(ctx context.Context, userID int64, name, avatar string) (domain.User, error) {
	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}
	if err = user.UpdateProfile(name, avatar); err != nil {
		return domain.User{}, err
	}
	if err = s.repository.UpdateUser(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}
