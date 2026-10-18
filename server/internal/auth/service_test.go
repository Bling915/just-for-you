package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"
)

type fakeExchange struct{ openid, key string }

func (f fakeExchange) Exchange(context.Context, string) (string, string, error) {
	return f.openid, f.key, nil
}

type memoryRepo struct {
	next   int64
	users  map[string]int64
	tokens map[[32]byte]RefreshRecord
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{users: map[string]int64{}, tokens: map[[32]byte]RefreshRecord{}}
}
func (r *memoryRepo) FindOrCreateUser(_ context.Context, openid string) (int64, error) {
	if id := r.users[openid]; id != 0 {
		return id, nil
	}
	r.next++
	r.users[openid] = r.next
	return r.next, nil
}
func (r *memoryRepo) StoreRefresh(_ context.Context, rec RefreshRecord) error {
	r.tokens[rec.Hash] = rec
	return nil
}
func (r *memoryRepo) RotateRefresh(_ context.Context, old [32]byte, next RefreshRecord, now time.Time) (int64, error) {
	rec, ok := r.tokens[old]
	if !ok || rec.RevokedAt != nil || !rec.ExpiresAt.After(now) {
		return 0, ErrInvalidRefresh
	}
	rec.RevokedAt = &now
	r.tokens[old] = rec
	next.UserID = rec.UserID
	r.tokens[next.Hash] = next
	return rec.UserID, nil
}
func (r *memoryRepo) RevokeRefresh(_ context.Context, hash [32]byte, now time.Time) error {
	rec, ok := r.tokens[hash]
	if !ok {
		return ErrInvalidRefresh
	}
	rec.RevokedAt = &now
	r.tokens[hash] = rec
	return nil
}

func TestLoginRejectsEmptyCode(t *testing.T) {
	svc := NewService(fakeExchange{openid: "same"}, newMemoryRepo(), NewTokenManager([]byte("12345678901234567890123456789012")), time.Now)
	if _, err := svc.Login(context.Background(), " "); !errors.Is(err, ErrCodeRequired) {
		t.Fatalf("error=%v", err)
	}
}

func TestLoginReusesUserForSameOpenID(t *testing.T) {
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	repo := newMemoryRepo()
	svc := NewService(fakeExchange{openid: "same", key: "do-not-log"}, repo, NewTokenManager([]byte("12345678901234567890123456789012")), func() time.Time { return now })
	a, _ := svc.Login(context.Background(), "first-code")
	b, _ := svc.Login(context.Background(), "second-code")
	if a.UserID != b.UserID || a.UserID == 0 {
		t.Fatalf("ids=%d,%d", a.UserID, b.UserID)
	}
	if !a.AccessExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("expiry=%s", a.AccessExpiresAt)
	}
	if _, err := svc.Tokens.ParseAccess(a.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.tokens[sha256.Sum256([]byte(a.RefreshToken))]; !ok {
		t.Fatal("refresh token hash not stored")
	}
}

func TestRefreshRejectsRevokedToken(t *testing.T) {
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	svc := NewService(fakeExchange{openid: "u"}, newMemoryRepo(), NewTokenManager([]byte("12345678901234567890123456789012")), func() time.Time { return now })
	s, _ := svc.Login(context.Background(), "code")
	_ = svc.Revoke(context.Background(), s.RefreshToken)
	if _, err := svc.Refresh(context.Background(), s.RefreshToken); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("error=%v", err)
	}
}
