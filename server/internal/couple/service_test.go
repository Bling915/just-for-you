package couple

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type memRepo struct {
	mu     sync.Mutex
	inv    Invite
	used   bool
	active map[int64]Couple
}

func (m *memRepo) SaveInvite(_ context.Context, i Invite) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inv = i
	return nil
}
func (m *memRepo) Preview(_ context.Context, c string) (Invite, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c != m.inv.Code {
		return Invite{}, ErrInviteInvalid
	}
	return m.inv, nil
}
func (m *memRepo) Accept(_ context.Context, u int64, c string, n time.Time) (Couple, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c != m.inv.Code {
		return Couple{}, ErrInviteInvalid
	}
	if m.used {
		return Couple{}, ErrInviteUsed
	}
	if !m.inv.ExpiresAt.After(n) {
		return Couple{}, ErrInviteExpired
	}
	if u == m.inv.CreatorUserID {
		return Couple{}, ErrCannotBindSelf
	}
	if _, ok := m.active[u]; ok {
		return Couple{}, ErrAlreadyBound
	}
	if _, ok := m.active[m.inv.CreatorUserID]; ok {
		return Couple{}, ErrAlreadyBound
	}
	x := Couple{ID: 1, UserAID: m.inv.CreatorUserID, UserBID: u, Status: "active", BoundAt: n}
	m.used = true
	m.active[u] = x
	m.active[m.inv.CreatorUserID] = x
	return x, nil
}
func (m *memRepo) GetActive(_ context.Context, u int64) (Couple, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	x, ok := m.active[u]
	if !ok {
		return Couple{}, ErrNoActiveCouple
	}
	return x, nil
}
func (m *memRepo) Unbind(_ context.Context, u int64, n time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	x, ok := m.active[u]
	if !ok {
		return ErrNoActiveCouple
	}
	delete(m.active, x.UserAID)
	delete(m.active, x.UserBID)
	return nil
}
func fixedCode() (string, error) { return "AB3K9Z", nil }
func TestCreateInviteHasSixCharactersAndTenMinuteExpiry(t *testing.T) {
	now := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	s := NewService(&memRepo{active: map[int64]Couple{}}, func() time.Time { return now }, fixedCode)
	x, e := s.CreateInvite(context.Background(), 7)
	if e != nil || x.Code != "AB3K9Z" || !x.ExpiresAt.Equal(now.Add(10*time.Minute)) {
		t.Fatalf("invite=%+v err=%v", x, e)
	}
}
func TestAcceptInviteRejectsSelf(t *testing.T) {
	now := time.Now()
	r := &memRepo{active: map[int64]Couple{}, inv: Invite{Code: "AB3K9Z", CreatorUserID: 7, ExpiresAt: now.Add(time.Minute)}}
	s := NewService(r, func() time.Time { return now }, fixedCode)
	if _, e := s.AcceptInvite(context.Background(), 7, "AB3K9Z"); !errors.Is(e, ErrCannotBindSelf) {
		t.Fatalf("err=%v", e)
	}
}
func TestConcurrentAcceptInviteOnlyOneSucceeds(t *testing.T) {
	now := time.Now()
	r := &memRepo{active: map[int64]Couple{}, inv: Invite{Code: "AB3K9Z", CreatorUserID: 1, ExpiresAt: now.Add(time.Minute)}}
	s := NewService(r, func() time.Time { return now }, fixedCode)
	var wg sync.WaitGroup
	success := 0
	var mu sync.Mutex
	for _, u := range []int64{2, 3} {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			if _, e := s.AcceptInvite(context.Background(), id, "AB3K9Z"); e == nil {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}(u)
	}
	wg.Wait()
	if success != 1 {
		t.Fatalf("success=%d", success)
	}
}
func TestUnbindRemovesActiveCouple(t *testing.T) {
	now := time.Now()
	r := &memRepo{active: map[int64]Couple{1: {ID: 1, UserAID: 1, UserBID: 2}}}
	r.active[2] = r.active[1]
	s := NewService(r, func() time.Time { return now }, fixedCode)
	_ = s.Unbind(context.Background(), 1)
	if _, e := s.GetActive(context.Background(), 2); !errors.Is(e, ErrNoActiveCouple) {
		t.Fatalf("err=%v", e)
	}
}
