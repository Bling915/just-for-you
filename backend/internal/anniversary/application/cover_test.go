package anniversary

import (
	"context"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type coverStorageFake struct {
	verifyErr error
}

func (coverStorageFake) SignUpload(
	context.Context,
	string,
	string,
	time.Time,
) (string, error) {
	return "https://upload", nil
}

func (fake coverStorageFake) VerifyUploaded(
	context.Context,
	string,
	string,
	int64,
) error {
	return fake.verifyErr
}

func (coverStorageFake) SignRead(
	_ context.Context,
	key string,
	_ time.Time,
) (string, error) {
	return "https://read/" + key, nil
}

func TestUpdateAnniversaryConsumesMatchingCoverGrant(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	item := domain.RestoreAnniversary(
		7,
		9,
		7,
		"相识",
		now,
		true,
		domain.CountModeElapsed,
		true,
		"old-cover",
		"",
		"none",
		now,
		now,
	)
	store := newMemoryAnniversaryStore()
	store.items = []domain.Anniversary{item}
	store.grant = CoverUploadGrant{
		ObjectKey:   "new-cover",
		OwnerID:     7,
		ResourceID:  7,
		ContentType: "image/jpeg",
		Purpose:     "anniversary_cover",
		ExpiresAt:   now.Add(time.Minute),
	}
	usecase := NewUpdateAnniversary(
		fakeAnniversaryCouple{relationship: ActiveCouple{ID: 9, UserAID: 7}},
		store,
		coverStorageFake{},
		fixedAnniversaryClock{now},
	)

	_, err := usecase.Execute(context.Background(), UpdateAnniversaryCommand{
		UserID: 7,
		ID:     7,
		Draft: domain.AnniversaryDraft{
			Title:         "相识",
			Date:          now,
			CountMode:     domain.CountModeElapsed,
			RepeatsYearly: true,
			Pinned:        true,
			CoverRef:      "new-cover",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if store.updatedWithKey != "new-cover" {
		t.Fatalf("updatedWithKey=%q", store.updatedWithKey)
	}
}
