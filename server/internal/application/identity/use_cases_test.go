package identity

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
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

func tokenGenerator(values ...string) func() (string, error) {
	index := 0
	return func() (string, error) {
		value := values[index]
		index++
		return value, nil
	}
}

func TestLoginExecuteReusesExistingUserAndIssuesSession(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	useCase := NewLoginUseCase(fakeIdentityProvider{ExternalIdentity{OpenID: "same"}}, repo, fakeTokens{}, fixedClock{now}, tokenGenerator("first", "second"))
	first, err := useCase.Execute(context.Background(), LoginCommand{Code: "one"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := useCase.Execute(context.Background(), LoginCommand{Code: "two"})
	if err != nil {
		t.Fatal(err)
	}
	if first.UserID != second.UserID || !first.AccessExpiresAt.Equal(now.Add(15*time.Minute)) {
		t.Fatalf("sessions = %+v %+v", first, second)
	}
	stored := repo.refresh[sha256.Sum256([]byte(first.RefreshToken))]
	if !stored.ExpiresAt().Equal(now.Add(30 * 24 * time.Hour)) {
		t.Fatalf("refresh expiry = %s", stored.ExpiresAt())
	}
}

func TestRefreshSessionExecuteRotatesRefreshToken(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	login := NewLoginUseCase(fakeIdentityProvider{ExternalIdentity{OpenID: "same"}}, repo, fakeTokens{}, fixedClock{now}, tokenGenerator("old"))
	session, err := login.Execute(context.Background(), LoginCommand{Code: "code"})
	if err != nil {
		t.Fatal(err)
	}
	refresh := NewRefreshSessionUseCase(repo, fakeTokens{}, fixedClock{now}, tokenGenerator("new", "unused"))
	next, err := refresh.Execute(context.Background(), RefreshSessionCommand{RefreshToken: session.RefreshToken})
	if err != nil || next.RefreshToken != "new" {
		t.Fatalf("session = %+v, error = %v", next, err)
	}
	if _, err := refresh.Execute(context.Background(), RefreshSessionCommand{RefreshToken: session.RefreshToken}); !errors.Is(err, domain.ErrInvalidRefreshToken) {
		t.Fatalf("reusing old token error = %v", err)
	}
}

func TestLogoutExecuteRevokesRefreshToken(t *testing.T) {
	now := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	repo.refresh[sha256.Sum256([]byte("token"))] = domain.NewRefreshToken(sha256.Sum256([]byte("token")), 7, now.Add(time.Hour))
	if err := NewLogoutUseCase(repo, fixedClock{now}).Execute(context.Background(), LogoutCommand{RefreshToken: "token"}); err != nil {
		t.Fatal(err)
	}
	if repo.refresh[sha256.Sum256([]byte("token"))].IsUsableAt(now.Add(time.Minute)) {
		t.Fatal("token remains usable after logout")
	}
}

func TestGetProfileExecuteReturnsUser(t *testing.T) {
	repo := newMemoryRepository()
	user, _ := repo.FindOrCreateUser(context.Background(), "open-id")
	got, err := NewGetProfileUseCase(repo).Execute(context.Background(), GetProfileQuery{UserID: user.ID()})
	if err != nil || got.ID() != user.ID() {
		t.Fatalf("user = %+v, error = %v", got, err)
	}
}

func TestUpdateProfileExecuteValidatesAndPersistsUser(t *testing.T) {
	repo := newMemoryRepository()
	user, _ := repo.FindOrCreateUser(context.Background(), "open-id")
	useCase := NewUpdateProfileUseCase(repo)
	if _, err := useCase.Execute(context.Background(), UpdateProfileCommand{UserID: user.ID(), DisplayName: " ", AvatarRef: "avatar"}); !errors.Is(err, domain.ErrInvalidNickname) {
		t.Fatalf("invalid name error = %v", err)
	}
	updated, err := useCase.Execute(context.Background(), UpdateProfileCommand{UserID: user.ID(), DisplayName: " 小南 ", AvatarRef: "avatar"})
	if err != nil || updated.DisplayName() != "小南" || repo.users["open-id"].DisplayName() != "小南" {
		t.Fatalf("updated = %+v, stored = %+v, error = %v", updated, repo.users["open-id"], err)
	}
}
