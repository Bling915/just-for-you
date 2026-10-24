package companionship

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/companionship/domain"
	coupledomain "github.com/nanfang/just-for-you/server/internal/couple/domain"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type memoryCouples struct {
	relationship ActiveCouple
	err          error
}

func (r memoryCouples) GetActiveCouple(context.Context, int64) (ActiveCouple, error) {
	return r.relationship, r.err
}

type memoryRepository struct {
	presences []domain.Presence
	nudges    []domain.Nudge
}

type memoryPreferences struct{ value NotificationPreferences }

func (m *memoryPreferences) GetPreferences(context.Context, int64, int64) (NotificationPreferences, error) {
	return m.value, nil
}
func (m *memoryPreferences) SavePreferences(_ context.Context, _, _ int64, value NotificationPreferences) (NotificationPreferences, error) {
	m.value = value
	return value, nil
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

func TestUpdatePresenceRequiresActiveCoupleAndStoresStatus(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	repo := &memoryRepository{}
	useCase := NewUpdatePresence(memoryCouples{relationship: ActiveCouple{ID: 9, UserAID: 7, UserBID: 8}}, repo, fixedClock{now})

	result, err := useCase.Execute(context.Background(), UpdatePresenceCommand{UserID: 7, Preset: "in_class", ExpiresAt: timePointer(now.Add(time.Hour))})
	if err != nil || result.UserID() != 7 || len(repo.presences) != 1 {
		t.Fatalf("result=%+v stored=%d err=%v", result, len(repo.presences), err)
	}

	ended := NewUpdatePresence(memoryCouples{err: coupledomain.ErrNoActiveCouple}, repo, fixedClock{now})
	if _, err = ended.Execute(context.Background(), UpdatePresenceCommand{UserID: 7, Preset: "tired"}); !errors.Is(err, coupledomain.ErrNoActiveCouple) {
		t.Fatalf("err=%v", err)
	}
}

func TestGetSnapshotHidesExpiredPresence(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	expires := now.Add(-time.Second)
	repo := &memoryRepository{presences: []domain.Presence{domain.RestorePresence(1, 9, 8, domain.PresetTired, "", &expires, now.Add(-time.Hour))}}
	useCase := NewGetSnapshot(memoryCouples{relationship: ActiveCouple{ID: 9, UserAID: 7, UserBID: 8}}, repo, repo, fixedClock{now})

	result, err := useCase.Execute(context.Background(), GetSnapshotQuery{UserID: 7})
	if err != nil || result.PartnerPresence != nil {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestSendNudgeTargetsPartnerAndThrottlesRapidRepeats(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	repo := &memoryRepository{}
	useCase := NewSendNudge(memoryCouples{relationship: ActiveCouple{ID: 9, UserAID: 7, UserBID: 8}}, repo, fixedClock{now})

	nudge, err := useCase.Execute(context.Background(), SendNudgeCommand{UserID: 7, Kind: "hug"})
	if err != nil || nudge.RecipientID() != 8 {
		t.Fatalf("nudge=%+v err=%v", nudge, err)
	}
	if _, err = useCase.Execute(context.Background(), SendNudgeCommand{UserID: 7, Kind: "hug"}); !errors.Is(err, domain.ErrNudgeThrottled) {
		t.Fatalf("err=%v", err)
	}
}

func TestNotificationPreferencesUseSeparateQueriesAndCommands(t *testing.T) {
	relationships := memoryCouples{relationship: ActiveCouple{ID: 9, UserAID: 7, UserBID: 8}}
	store := &memoryPreferences{}
	available := map[string]bool{"moment": true, "presence": false, "nudge": true}
	update := NewUpdateNotificationPreferences(relationships, store, available)
	get := NewGetNotificationPreferences(relationships, store, available)

	_, err := update.Execute(context.Background(), UpdateNotificationPreferencesCommand{UserID: 7, Preferences: NotificationPreferences{MomentEnabled: true, PresenceEnabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := get.Execute(context.Background(), GetNotificationPreferencesQuery{UserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if !result.MomentEnabled || result.PresenceEnabled || !result.Available["nudge"] {
		t.Fatalf("result=%+v", result)
	}
}

func timePointer(value time.Time) *time.Time { return &value }
