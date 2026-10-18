package couple

import (
	"context"
	"crypto/rand"
	"strings"
	"time"
)

type Repository interface {
	SaveInvite(context.Context, Invite) error
	Preview(context.Context, string) (Invite, error)
	Accept(context.Context, int64, string, time.Time) (Couple, error)
	GetActive(context.Context, int64) (Couple, error)
	Unbind(context.Context, int64, time.Time) error
}
type Service struct {
	repo Repository
	now  func() time.Time
	code func() (string, error)
}

func NewService(r Repository, n func() time.Time, g func() (string, error)) *Service {
	return &Service{r, n, g}
}
func RandomCode() (string, error) {
	const chars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, 6)
	raw := make([]byte, 6)
	if _, e := rand.Read(raw); e != nil {
		return "", e
	}
	for i := range b {
		b[i] = chars[int(raw[i])%len(chars)]
	}
	return string(b), nil
}
func (s *Service) CreateInvite(ctx context.Context, u int64) (Invite, error) {
	c, e := s.code()
	if e != nil {
		return Invite{}, e
	}
	x := Invite{Code: c, CreatorUserID: u, ExpiresAt: s.now().Add(10 * time.Minute)}
	return x, s.repo.SaveInvite(ctx, x)
}
func (s *Service) PreviewInvite(ctx context.Context, u int64, c string) (InvitePreview, error) {
	x, e := s.repo.Preview(ctx, strings.ToUpper(strings.TrimSpace(c)))
	return InvitePreview{Code: x.Code, CreatorUserID: x.CreatorUserID, ExpiresAt: x.ExpiresAt}, e
}
func (s *Service) AcceptInvite(ctx context.Context, u int64, c string) (Couple, error) {
	return s.repo.Accept(ctx, u, strings.ToUpper(strings.TrimSpace(c)), s.now())
}
func (s *Service) GetActive(ctx context.Context, u int64) (Couple, error) {
	return s.repo.GetActive(ctx, u)
}
func (s *Service) Unbind(ctx context.Context, u int64) error { return s.repo.Unbind(ctx, u, s.now()) }
