package application

import (
	"context"
	"errors"
	"github.com/nanfang/just-for-you/server/internal/relationship/domain"
	"sync"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type memoryRepository struct {
	mu         sync.Mutex
	invitation domain.Invitation
	active     map[int64]domain.Relationship
}

func (r *memoryRepository) SaveInvitation(_ context.Context, i domain.Invitation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invitation = i
	return nil
}
func (r *memoryRepository) FindInvitation(_ context.Context, code string) (domain.Invitation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.invitation.Code() != code {
		return domain.Invitation{}, domain.ErrInvitationInvalid
	}
	return r.invitation, nil
}
func (r *memoryRepository) AcceptInvitation(_ context.Context, code string, userID int64, now time.Time) (domain.Relationship, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.invitation.Code() != code {
		return domain.Relationship{}, domain.ErrInvitationInvalid
	}
	if err := r.invitation.CanBeAcceptedBy(userID, now); err != nil {
		return domain.Relationship{}, err
	}
	if _, ok := r.active[userID]; ok {
		return domain.Relationship{}, domain.ErrAlreadyBound
	}
	if _, ok := r.active[r.invitation.CreatorID()]; ok {
		return domain.Relationship{}, domain.ErrAlreadyBound
	}
	relationship := domain.NewRelationship(1, r.invitation.CreatorID(), userID, now)
	r.invitation.MarkUsed(now, userID)
	r.active[userID] = relationship
	r.active[r.invitation.CreatorID()] = relationship
	return relationship, nil
}
func (r *memoryRepository) GetActive(_ context.Context, userID int64) (domain.Relationship, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	x, ok := r.active[userID]
	if !ok {
		return domain.Relationship{}, domain.ErrNoActiveRelationship
	}
	return x, nil
}
func (r *memoryRepository) End(_ context.Context, userID int64, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	x, ok := r.active[userID]
	if !ok {
		return domain.ErrNoActiveRelationship
	}
	delete(r.active, x.UserAID())
	delete(r.active, x.UserBID())
	return nil
}
func TestCreateInvitationExpiresInTenMinutes(t *testing.T) {
	now := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	repo := &memoryRepository{active: map[int64]domain.Relationship{}}
	service := NewService(repo, fixedClock{now}, func() (string, error) { return "AB3K9Z", nil })
	invitation, err := service.CreateInvitation(context.Background(), 1)
	if err != nil || !invitation.ExpiresAt().Equal(now.Add(10*time.Minute)) {
		t.Fatalf("invitation=%+v err=%v", invitation, err)
	}
}
func TestConcurrentBindOnlyOneSucceeds(t *testing.T) {
	now := time.Now()
	invitation, _ := domain.NewInvitation("AB3K9Z", 1, now.Add(time.Minute))
	repo := &memoryRepository{invitation: invitation, active: map[int64]domain.Relationship{}}
	service := NewService(repo, fixedClock{now}, func() (string, error) { return "AB3K9Z", nil })
	var wg sync.WaitGroup
	success := 0
	var mu sync.Mutex
	for _, id := range []int64{2, 3} {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			if _, err := service.BindRelationship(context.Background(), userID, "AB3K9Z"); err == nil {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}(id)
	}
	wg.Wait()
	if success != 1 {
		t.Fatalf("success=%d", success)
	}
}
func TestEndRelationshipRemovesAccess(t *testing.T) {
	now := time.Now()
	relationship := domain.NewRelationship(1, 1, 2, now)
	repo := &memoryRepository{active: map[int64]domain.Relationship{1: relationship, 2: relationship}}
	service := NewService(repo, fixedClock{now}, func() (string, error) { return "AB3K9Z", nil })
	if err := service.EndRelationship(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetRelationship(context.Background(), 2); !errors.Is(err, domain.ErrNoActiveRelationship) {
		t.Fatalf("err=%v", err)
	}
}
