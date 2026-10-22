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
	grants  map[string]AvatarGrant
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{users: map[string]domain.User{}, refresh: map[[32]byte]domain.RefreshToken{}, grants: map[string]AvatarGrant{}}
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
func (r *memoryRepository) CreateAvatarGrant(_ context.Context, grant AvatarGrant) error {
	r.grants[grant.ObjectKey] = grant
	return nil
}
func (r *memoryRepository) GetAvatarGrant(_ context.Context, key string) (AvatarGrant, error) {
	g, ok := r.grants[key]
	if !ok {
		return AvatarGrant{}, domain.ErrInvalidAvatar
	}
	return g, nil
}
func (r *memoryRepository) UpdateUserWithAvatarGrant(_ context.Context, user domain.User, key string, now time.Time) error {
	g, ok := r.grants[key]
	if !ok || g.ConsumedAt != nil || !g.ExpiresAt.After(now) || g.OwnerID != user.ID() || g.Purpose != "avatar" {
		return domain.ErrInvalidAvatar
	}
	g.ConsumedAt = &now
	r.grants[key] = g
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

type fakeAvatarStorage struct{ verifyErr error }

func (fakeAvatarStorage) SignUpload(_ context.Context, key, contentType string, _ time.Time) (string, error) {
	return "https://upload/" + key, nil
}
func (f fakeAvatarStorage) VerifyUploaded(context.Context, string, string, int64) error {
	return f.verifyErr
}
func (fakeAvatarStorage) SignRead(_ context.Context, key string, _ time.Time) (string, error) {
	return "https://read/" + key, nil
}

func TestGetProfileExecuteReturnsSignedAvatarURL(t *testing.T) {
	repo := newMemoryRepository()
	user, _ := repo.FindOrCreateUser(context.Background(), "open-id")
	_ = user.UpdateProfile("小南", "users/1/avatars/a.jpg")
	_ = repo.UpdateUser(context.Background(), user)
	got, err := NewGetProfileUseCase(repo, fakeAvatarStorage{}, fixedClock{time.Now()}).Execute(context.Background(), GetProfileQuery{UserID: user.ID()})
	if err != nil || got.User.ID() != user.ID() || got.AvatarURL != "https://read/users/1/avatars/a.jpg" {
		t.Fatalf("user = %+v, error = %v", got, err)
	}
}

func TestUpdateProfileExecuteValidatesAndPersistsUser(t *testing.T) {
	repo := newMemoryRepository()
	user, _ := repo.FindOrCreateUser(context.Background(), "open-id")
	useCase := NewUpdateProfileUseCase(repo, fakeAvatarStorage{}, fixedClock{time.Now()})
	if _, err := useCase.Execute(context.Background(), UpdateProfileCommand{UserID: user.ID(), DisplayName: " ", AvatarRef: "avatar"}); !errors.Is(err, domain.ErrInvalidNickname) {
		t.Fatalf("invalid name error = %v", err)
	}
	updated, err := useCase.Execute(context.Background(), UpdateProfileCommand{UserID: user.ID(), DisplayName: " 小南 ", AvatarRef: ""})
	if err != nil || updated.DisplayName() != "小南" || repo.users["open-id"].DisplayName() != "小南" {
		t.Fatalf("updated = %+v, stored = %+v, error = %v", updated, repo.users["open-id"], err)
	}
}

func TestRequestAvatarUploadAcceptsImagesUpToFiveMiB(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	u := NewRequestAvatarUploadUseCase(repo, fakeAvatarStorage{}, fixedClock{now}, func() string { return "fixed" })
	for _, kind := range []string{"image/jpeg", "image/png", "image/webp"} {
		g, err := u.Execute(context.Background(), RequestAvatarUploadCommand{UserID: 7, ContentType: kind, Size: 5 << 20})
		if err != nil || g.ObjectKey != "users/7/avatars/fixed" || g.UploadURL == "" {
			t.Fatalf("kind=%s grant=%+v err=%v", kind, g, err)
		}
	}
	if _, err := u.Execute(context.Background(), RequestAvatarUploadCommand{UserID: 7, ContentType: "image/gif", Size: 100}); !errors.Is(err, domain.ErrInvalidAvatar) {
		t.Fatalf("gif err=%v", err)
	}
	if _, err := u.Execute(context.Background(), RequestAvatarUploadCommand{UserID: 7, ContentType: "image/jpeg", Size: (5 << 20) + 1}); !errors.Is(err, domain.ErrInvalidAvatar) {
		t.Fatalf("size err=%v", err)
	}
}

func TestUpdateProfileRequiresVerifiedUsableOwnedAvatarGrant(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	user, _ := repo.FindOrCreateUser(context.Background(), "open-id")
	_ = user.UpdateProfile("旧昵称", "users/1/avatars/old.jpg")
	_ = repo.UpdateUser(context.Background(), user)
	valid := AvatarGrant{ObjectKey: "users/1/avatars/new.jpg", OwnerID: 1, ContentType: "image/jpeg", Purpose: "avatar", ExpiresAt: now.Add(time.Minute)}
	cases := []struct {
		name   string
		grant  AvatarGrant
		key    string
		verify error
	}{
		{"wrong prefix", valid, "users/2/avatars/new.jpg", nil},
		{"wrong purpose", func() AvatarGrant { g := valid; g.Purpose = "moment_photo"; return g }(), valid.ObjectKey, nil},
		{"expired", func() AvatarGrant { g := valid; g.ExpiresAt = now.Add(-time.Second); return g }(), valid.ObjectKey, nil},
		{"consumed", func() AvatarGrant { g := valid; x := now; g.ConsumedAt = &x; return g }(), valid.ObjectKey, nil},
		{"not uploaded", valid, valid.ObjectKey, errors.New("missing")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo.grants = map[string]AvatarGrant{tc.grant.ObjectKey: tc.grant}
			u := NewUpdateProfileUseCase(repo, fakeAvatarStorage{verifyErr: tc.verify}, fixedClock{now})
			_, err := u.Execute(context.Background(), UpdateProfileCommand{UserID: 1, DisplayName: "新昵称", AvatarRef: tc.key})
			if !errors.Is(err, domain.ErrInvalidAvatar) {
				t.Fatalf("err=%v", err)
			}
			stored, _ := repo.GetUser(context.Background(), 1)
			if stored.AvatarRef() != "users/1/avatars/old.jpg" {
				t.Fatalf("avatar changed to %q", stored.AvatarRef())
			}
		})
	}
	repo.grants = map[string]AvatarGrant{valid.ObjectKey: valid}
	u := NewUpdateProfileUseCase(repo, fakeAvatarStorage{}, fixedClock{now})
	updated, err := u.Execute(context.Background(), UpdateProfileCommand{UserID: 1, DisplayName: "新昵称", AvatarRef: valid.ObjectKey})
	if err != nil || updated.AvatarRef() != valid.ObjectKey || repo.grants[valid.ObjectKey].ConsumedAt == nil {
		t.Fatalf("updated=%+v grant=%+v err=%v", updated, repo.grants[valid.ObjectKey], err)
	}
}
