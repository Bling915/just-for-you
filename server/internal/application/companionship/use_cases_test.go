package companionship

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type memoryRelationships struct {
	relationship ActiveRelationship
	err          error
}

func (r memoryRelationships) GetActiveRelationship(context.Context, int64) (ActiveRelationship, error) {
	return r.relationship, r.err
}

type memoryRepository struct {
	presences []domain.Presence
	nudges    []domain.Nudge
}

func (r *memoryRepository) SavePresence(_ context.Context, presence domain.Presence) (domain.Presence, error) {
	r.presences = append(r.presences, presence)
	return presence, nil
}
func (r *memoryRepository) LatestPresence(_ context.Context, coupleID, userID int64) (*domain.Presence, error) {
	for i := len(r.presences) - 1; i >= 0; i-- {
		if r.presences[i].CoupleID() == coupleID && r.presences[i].UserID() == userID {
			p := r.presences[i]
			return &p, nil
		}
	}
	return nil, nil
}
func (r *memoryRepository) LastNudgeAt(_ context.Context, coupleID, senderID int64) (*time.Time, error) {
	for i := len(r.nudges) - 1; i >= 0; i-- {
		if r.nudges[i].CoupleID() == coupleID && r.nudges[i].SenderID() == senderID {
			at := r.nudges[i].CreatedAt()
			return &at, nil
		}
	}
	return nil, nil
}
func (r *memoryRepository) SaveNudge(_ context.Context, nudge domain.Nudge) (domain.Nudge, error) {
	r.nudges = append(r.nudges, nudge)
	return nudge, nil
}
func (r *memoryRepository) LatestNudge(_ context.Context, coupleID int64) (*domain.Nudge, error) {
	if len(r.nudges) == 0 {
		return nil, nil
	}
	n := r.nudges[len(r.nudges)-1]
	return &n, nil
}

func TestUpdatePresenceRequiresActiveRelationshipAndStoresStatus(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	repo := &memoryRepository{}
	useCase := NewUpdatePresenceUseCase(memoryRelationships{relationship: ActiveRelationship{ID: 9, UserAID: 7, UserBID: 8}}, repo, fixedClock{now})

	result, err := useCase.Execute(context.Background(), UpdatePresenceCommand{UserID: 7, Preset: "in_class", ExpiresAt: timePointer(now.Add(time.Hour))})
	if err != nil || result.UserID() != 7 || len(repo.presences) != 1 {
		t.Fatalf("result=%+v stored=%d err=%v", result, len(repo.presences), err)
	}

	ended := NewUpdatePresenceUseCase(memoryRelationships{err: relationshipdomain.ErrNoActiveRelationship}, repo, fixedClock{now})
	if _, err = ended.Execute(context.Background(), UpdatePresenceCommand{UserID: 7, Preset: "tired"}); !errors.Is(err, relationshipdomain.ErrNoActiveRelationship) {
		t.Fatalf("err=%v", err)
	}
}

func TestGetSnapshotHidesExpiredPresence(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	expires := now.Add(-time.Second)
	repo := &memoryRepository{presences: []domain.Presence{domain.RestorePresence(1, 9, 8, domain.PresetTired, "", &expires, now.Add(-time.Hour))}}
	useCase := NewGetSnapshotUseCase(memoryRelationships{relationship: ActiveRelationship{ID: 9, UserAID: 7, UserBID: 8}}, repo, fixedClock{now})

	result, err := useCase.Execute(context.Background(), GetSnapshotQuery{UserID: 7})
	if err != nil || result.PartnerPresence != nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestSendNudgeTargetsPartnerAndThrottlesRapidRepeats(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	repo := &memoryRepository{}
	useCase := NewSendNudgeUseCase(memoryRelationships{relationship: ActiveRelationship{ID: 9, UserAID: 7, UserBID: 8}}, repo, fixedClock{now})

	nudge, err := useCase.Execute(context.Background(), SendNudgeCommand{UserID: 7, Kind: "hug"})
	if err != nil || nudge.RecipientID() != 8 {
		t.Fatalf("nudge=%+v err=%v", nudge, err)
	}
	if _, err = useCase.Execute(context.Background(), SendNudgeCommand{UserID: 7, Kind: "hug"}); !errors.Is(err, domain.ErrNudgeThrottled) {
		t.Fatalf("err=%v", err)
	}
}

func timePointer(value time.Time) *time.Time { return &value }
