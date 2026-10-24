package application

import (
	"context"
	"errors"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	"sync"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type memoryRepository struct {
	mu         sync.Mutex
	invitation domain.Invitation
	active     map[int64]domain.Couple
}

func (r *memoryRepository) CreateSoloCouple(_ context.Context, userID int64, now time.Time) (domain.Couple, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.active[userID]; ok {
		return domain.Couple{}, domain.ErrAlreadyBound
	}
	partnerID := userID + 1000
	relationship := domain.NewCouple(2, userID, partnerID, now)
	r.active[userID] = relationship
	r.active[partnerID] = relationship
	return relationship, nil
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
func (r *memoryRepository) AcceptInvitation(_ context.Context, code string, userID int64, now time.Time) (domain.Couple, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.invitation.Code() != code {
		return domain.Couple{}, domain.ErrInvitationInvalid
	}
	if err := r.invitation.CanBeAcceptedBy(userID, now); err != nil {
		return domain.Couple{}, err
	}
	if _, ok := r.active[userID]; ok {
		return domain.Couple{}, domain.ErrAlreadyBound
	}
	if _, ok := r.active[r.invitation.CreatorID()]; ok {
		return domain.Couple{}, domain.ErrAlreadyBound
	}
	relationship := domain.NewCouple(1, r.invitation.CreatorID(), userID, now)
	r.invitation.MarkUsed(now, userID)
	r.active[userID] = relationship
	r.active[r.invitation.CreatorID()] = relationship
	return relationship, nil
}
func (r *memoryRepository) GetActive(_ context.Context, userID int64) (domain.Couple, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	x, ok := r.active[userID]
	if !ok {
		return domain.Couple{}, domain.ErrNoActiveCouple
	}
	return x, nil
}
func (r *memoryRepository) GetActiveDetails(ctx context.Context, userID int64) (domain.Couple, PartnerProfile, bool, error) {
	relationship, err := r.GetActive(ctx, userID)
	if err != nil {
		return domain.Couple{}, PartnerProfile{}, false, err
	}
	partnerID := relationship.UserAID()
	if partnerID == userID {
		partnerID = relationship.UserBID()
	}
	if partnerID >= 1000 {
		return relationship, PartnerProfile{UserID: partnerID, DisplayName: "测试搭子"}, true, nil
	}
	return relationship, PartnerProfile{UserID: partnerID, DisplayName: "小南", AvatarRef: "cos://avatar"}, false, nil
}
func (r *memoryRepository) End(_ context.Context, userID int64, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	x, ok := r.active[userID]
	if !ok {
		return domain.ErrNoActiveCouple
	}
	delete(r.active, x.UserAID())
	delete(r.active, x.UserBID())
	return nil
}
func TestCreateInvitationExpiresInTenMinutes(t *testing.T) {
	now := time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)
	repo := &memoryRepository{active: map[int64]domain.Couple{}}
	useCase := NewCreateInvitation(repo, fixedClock{now}, func() (string, error) { return "AB3K9Z", nil })
	invitation, err := useCase.Execute(context.Background(), CreateInvitationCommand{UserID: 1})
	if err != nil || !invitation.ExpiresAt().Equal(now.Add(10*time.Minute)) {
		t.Fatalf("invitation=%+v err=%v", invitation, err)
	}
}
func TestConcurrentBindOnlyOneSucceeds(t *testing.T) {
	now := time.Now()
	invitation, _ := domain.NewInvitation("AB3K9Z", 1, now.Add(time.Minute))
	repo := &memoryRepository{invitation: invitation, active: map[int64]domain.Couple{}}
	useCase := NewBindCouple(repo, fixedClock{now})
	var wg sync.WaitGroup
	success := 0
	var mu sync.Mutex
	for _, id := range []int64{2, 3} {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			if _, err := useCase.Execute(context.Background(), BindCoupleCommand{UserID: userID, Code: "AB3K9Z"}); err == nil {
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
func TestEndCoupleRemovesAccess(t *testing.T) {
	now := time.Now()
	relationship := domain.NewCouple(1, 1, 2, now)
	repo := &memoryRepository{active: map[int64]domain.Couple{1: relationship, 2: relationship}}
	end := NewEndCouple(repo, fixedClock{now})
	get := NewGetCurrentCouple(repo)
	if err := end.Execute(context.Background(), EndCoupleCommand{UserID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := get.Execute(context.Background(), GetCoupleQuery{UserID: 2}); !errors.Is(err, domain.ErrNoActiveCouple) {
		t.Fatalf("err=%v", err)
	}
}

func TestEnableSoloTestCreatesCoupleForCurrentUser(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	repo := &memoryRepository{active: map[int64]domain.Couple{}}
	useCase := NewEnableSoloTest(repo, fixedClock{now})

	relationship, err := useCase.Execute(context.Background(), EnableSoloTestCommand{UserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if relationship.UserAID() != 7 || relationship.UserBID() != 1007 || !relationship.BoundAt().Equal(now) {
		t.Fatalf("relationship=%+v", relationship)
	}
}

func TestEnableSoloTestRejectsAlreadyBoundUser(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	existing := domain.NewCouple(1, 7, 8, now)
	repo := &memoryRepository{active: map[int64]domain.Couple{7: existing, 8: existing}}
	useCase := NewEnableSoloTest(repo, fixedClock{now})

	_, err := useCase.Execute(context.Background(), EnableSoloTestCommand{UserID: 7})
	if !errors.Is(err, domain.ErrAlreadyBound) {
		t.Fatalf("err=%v", err)
	}
}

func TestGetCoupleReturnsPartnerProfileWithoutPrivateIdentity(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	relationship := domain.NewCouple(9, 7, 8, now)
	repo := &memoryRepository{active: map[int64]domain.Couple{7: relationship, 8: relationship}}
	useCase := NewGetCurrentCouple(repo)

	result, err := useCase.Execute(context.Background(), GetCoupleQuery{UserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if result.Partner.UserID != 8 || result.Partner.DisplayName != "小南" || result.Partner.AvatarURL != "cos://avatar" {
		t.Fatalf("partner=%+v", result.Partner)
	}
	if result.SoloTest {
		t.Fatal("real relationship reported as solo test")
	}
}

func TestGetCoupleMarksSoloTestPartner(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	relationship := domain.NewCouple(10, 7, 1007, now)
	repo := &memoryRepository{active: map[int64]domain.Couple{7: relationship, 1007: relationship}}
	useCase := NewGetCurrentCouple(repo)

	result, err := useCase.Execute(context.Background(), GetCoupleQuery{UserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if !result.SoloTest {
		t.Fatal("solo relationship was not identified")
	}
}

type avatarSignerStub struct{}

func (avatarSignerStub) SignRead(_ context.Context, key string, _ time.Time) (string, error) {
	return "https://signed/" + key, nil
}

func TestGetCoupleSignsPartnerAvatarReference(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	relationship := domain.NewCouple(9, 7, 8, now)
	repo := &memoryRepository{active: map[int64]domain.Couple{7: relationship, 8: relationship}}
	useCase := NewGetCurrentCoupleWithAvatars(repo, avatarSignerStub{}, fixedClock{now})
	result, err := useCase.Execute(context.Background(), GetCoupleQuery{UserID: 7})
	if err != nil || result.Partner.AvatarURL != "https://signed/cos://avatar" {
		t.Fatalf("partner=%+v err=%v", result.Partner, err)
	}
}
