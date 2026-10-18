package application

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/nanfang/just-for-you/server/internal/identity/domain"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fakeIdentityProvider struct{ identity ExternalIdentity }

func (f fakeIdentityProvider) Exchange(context.Context, string) (ExternalIdentity, error) {
	return f.identity, nil
}

type fakeTokens struct{}

func (fakeTokens) IssueAccess(userID int64, now time.Time) (string, time.Time, error) {
	return "access", now.Add(15 * time.Minute), nil
}

type memoryRepository struct {
	nextID  int64
	users   map[string]domain.User
	refresh map[[32]byte]domain.RefreshToken
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{users: map[string]domain.User{}, refresh: map[[32]byte]domain.RefreshToken{}}
}
func (r *memoryRepository) FindOrCreateUser(_ context.Context, openID string) (domain.User, error) {
	if user, ok := r.users[openID]; ok {
		return user, nil
	}
	r.nextID++
	user := domain.NewUser(r.nextID, openID)
	r.users[openID] = user
	return user, nil
}
func (r *memoryRepository) GetUser(_ context.Context, id int64) (domain.User, error) {
	for _, user := range r.users {
		if user.ID() == id {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}
func (r *memoryRepository) UpdateUser(_ context.Context, user domain.User) error {
	r.users[user.OpenID()] = user
	return nil
}
func (r *memoryRepository) StoreRefresh(_ context.Context, token domain.RefreshToken) error {
	r.refresh[token.Hash()] = token
	return nil
}
func (r *memoryRepository) RotateRefresh(_ context.Context, old [32]byte, next domain.RefreshToken, now time.Time) (int64, error) {
	current, ok := r.refresh[old]
	if !ok || !current.IsUsableAt(now) {
		return 0, domain.ErrInvalidRefreshToken
	}
	current.Revoke(now)
	r.refresh[old] = current
	next = domain.NewRefreshToken(next.Hash(), current.UserID(), next.ExpiresAt())
	r.refresh[next.Hash()] = next
	return current.UserID(), nil
}
func (r *memoryRepository) RevokeRefresh(_ context.Context, hash [32]byte, now time.Time) error {
	token, ok := r.refresh[hash]
	if !ok || !token.IsUsableAt(now) {
		return domain.ErrInvalidRefreshToken
	}
	token.Revoke(now)
	r.refresh[hash] = token
	return nil
}

func newService(now time.Time, repo *memoryRepository) *Service {
	return NewService(fakeIdentityProvider{ExternalIdentity{OpenID: "same", SessionKey: "secret"}}, repo, fakeTokens{}, fixedClock{now}, func() (string, error) { return "refresh-value", nil })
}

func TestLoginWithWeChatRejectsEmptyCode(t *testing.T) {
	_, err := newService(time.Now(), newMemoryRepository()).LoginWithWeChat(context.Background(), " ")
	if !errors.Is(err, domain.ErrCodeRequired) {
		t.Fatalf("error = %v", err)
	}
}

func TestLoginWithWeChatReusesUserAndSetsExpirations(t *testing.T) {
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	service := newService(now, repo)
	a, err := service.LoginWithWeChat(context.Background(), "one")
	if err != nil {
		t.Fatal(err)
	}
	b, err := service.LoginWithWeChat(context.Background(), "two")
	if err != nil {
		t.Fatal(err)
	}
	if a.UserID != b.UserID || !a.AccessExpiresAt.Equal(now.Add(15*time.Minute)) {
		t.Fatalf("sessions = %+v %+v", a, b)
	}
	token := repo.refresh[sha256.Sum256([]byte(a.RefreshToken))]
	if !token.ExpiresAt().Equal(now.Add(30 * 24 * time.Hour)) {
		t.Fatalf("refresh expiry = %s", token.ExpiresAt())
	}
}

func TestRefreshSessionRotatesAndRejectsRevokedToken(t *testing.T) {
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	service := newService(now, repo)
	session, _ := service.LoginWithWeChat(context.Background(), "code")
	if err := service.Logout(context.Background(), session.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RefreshSession(context.Background(), session.RefreshToken); !errors.Is(err, domain.ErrInvalidRefreshToken) {
		t.Fatalf("error = %v", err)
	}
}
