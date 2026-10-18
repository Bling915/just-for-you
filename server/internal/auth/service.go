package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"time"
)

type WeChatExchanger interface {
	Exchange(context.Context, string) (string, string, error)
}
type Repository interface {
	FindOrCreateUser(context.Context, string) (int64, error)
	StoreRefresh(context.Context, RefreshRecord) error
	RotateRefresh(context.Context, [32]byte, RefreshRecord, time.Time) (int64, error)
	RevokeRefresh(context.Context, [32]byte, time.Time) error
}
type Service struct {
	exchanger WeChatExchanger
	repo      Repository
	Tokens    TokenManager
	now       func() time.Time
}

func NewService(e WeChatExchanger, r Repository, t TokenManager, now func() time.Time) *Service {
	return &Service{e, r, t, now}
}
func (s *Service) Login(ctx context.Context, code string) (Session, error) {
	if strings.TrimSpace(code) == "" {
		return Session{}, ErrCodeRequired
	}
	openid, _, err := s.exchanger.Exchange(ctx, code)
	if err != nil {
		return Session{}, err
	}
	uid, err := s.repo.FindOrCreateUser(ctx, openid)
	if err != nil {
		return Session{}, err
	}
	return s.newSession(ctx, uid, nil)
}
func (s *Service) newSession(ctx context.Context, uid int64, old *[32]byte) (Session, error) {
	now := s.now()
	access, exp, err := s.Tokens.IssueAccess(uid, now)
	if err != nil {
		return Session{}, err
	}
	rawBytes := make([]byte, 32)
	if _, err = rand.Read(rawBytes); err != nil {
		return Session{}, err
	}
	raw := base64.RawURLEncoding.EncodeToString(rawBytes)
	rec := RefreshRecord{Hash: sha256.Sum256([]byte(raw)), UserID: uid, ExpiresAt: now.Add(30 * 24 * time.Hour)}
	if old == nil {
		err = s.repo.StoreRefresh(ctx, rec)
	} else {
		uid, err = s.repo.RotateRefresh(ctx, *old, rec, now)
		if err == nil {
			rec.UserID = uid
			access, exp, err = s.Tokens.IssueAccess(uid, now)
		}
	}
	if err != nil {
		return Session{}, err
	}
	return Session{uid, access, raw, exp}, nil
}
func (s *Service) Refresh(ctx context.Context, raw string) (Session, error) {
	h := sha256.Sum256([]byte(raw))
	return s.newSession(ctx, 0, &h)
}
func (s *Service) Revoke(ctx context.Context, raw string) error {
	h := sha256.Sum256([]byte(raw))
	return s.repo.RevokeRefresh(ctx, h, s.now())
}
