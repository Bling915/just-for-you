package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewMomentRejectsEmptyContent(t *testing.T) {
	_, err := NewMoment(1, 2, "", StatusNone, "", nil, time.Now())
	if !errors.Is(err, ErrEmptyMoment) {
		t.Fatalf("err=%v", err)
	}
}
func TestNewMomentEnforcesTextBoundary(t *testing.T) {
	if _, err := NewMoment(1, 2, strings.Repeat("界", 500), StatusNone, "", nil, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := NewMoment(1, 2, strings.Repeat("界", 501), StatusNone, "", nil, time.Now()); !errors.Is(err, ErrTextTooLong) {
		t.Fatalf("err=%v", err)
	}
}
func TestNewMomentEnforcesPlaceBoundary(t *testing.T) {
	if _, err := NewMoment(1, 2, "", StatusNone, strings.Repeat("界", 80), nil, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := NewMoment(1, 2, "", StatusNone, strings.Repeat("界", 81), nil, time.Now()); !errors.Is(err, ErrPlaceTooLong) {
		t.Fatalf("err=%v", err)
	}
}
func TestNewMomentValidatesStatus(t *testing.T) {
	for _, status := range []Status{StatusInClass, StatusMissYou, StatusTired} {
		if _, err := NewMoment(1, 2, "", status, "", nil, time.Now()); err != nil {
			t.Fatalf("status=%q err=%v", status, err)
		}
	}
	if _, err := NewMoment(1, 2, "", Status("unknown"), "", nil, time.Now()); !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("err=%v", err)
	}
}
func TestNewMomentKeepsOneToNinePhotosInOrder(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}
	moment, err := NewMoment(1, 2, "", StatusNone, "", keys, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for index, photo := range moment.Photos() {
		if photo.ObjectKey() != keys[index] || photo.Order() != index+1 {
			t.Fatalf("photo %d=%+v", index, photo)
		}
	}
	if _, err := NewMoment(1, 2, "", StatusNone, "", append(keys, "j"), time.Now()); !errors.Is(err, ErrTooManyPhotos) {
		t.Fatalf("err=%v", err)
	}
}
func TestNewMomentRejectsDuplicatePhotoKeys(t *testing.T) {
	if _, err := NewMoment(1, 2, "", StatusNone, "", []string{"same", "same"}, time.Now()); !errors.Is(err, ErrDuplicatePhoto) {
		t.Fatalf("err=%v", err)
	}
}

func TestMomentUpdateRequiresAuthorAndPreservesIdentity(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	moment := RestoreMomentDetails(RestoreInput{ID: 5, AuthorID: 7, RelationshipID: 9, Text: "旧内容", Status: StatusNone, Photos: []Photo{NewPhoto("old", 1)}, OccurredAt: now, CreatedAt: now, UpdatedAt: now, SpaceType: SpaceShared, Visibility: VisibilityCouple, Kind: KindStandard})
	occurred := now.Add(-time.Hour)
	latitude, longitude := 31.2304, 121.4737

	if _, err := moment.Update(8, UpdateInput{Text: "新内容", OccurredAt: occurred}, now); !errors.Is(err, ErrNotAuthor) {
		t.Fatalf("non-author err=%v", err)
	}
	updated, err := moment.Update(7, UpdateInput{Text: "新内容", Status: StatusMissYou, Place: "外滩", Latitude: &latitude, Longitude: &longitude, PhotoKeys: []string{"new-b", "new-a"}, OccurredAt: occurred}, now)
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID() != 5 || updated.AuthorID() != 7 || updated.RelationshipID() != 9 {
		t.Fatalf("identity changed: %+v", updated)
	}
	if !updated.OccurredAt().Equal(occurred) || updated.Latitude() == nil || updated.Longitude() == nil || updated.Photos()[0].ObjectKey() != "new-b" || updated.Photos()[1].Order() != 2 {
		t.Fatalf("updated=%+v photos=%+v", updated, updated.Photos())
	}
}

func TestMomentUpdateValidatesOccurrenceAndCoordinates(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	moment, _ := NewMoment(7, 9, "内容", StatusNone, "", nil, now)
	validLatitude, invalidLatitude, longitude := 31.0, 91.0, 121.0
	cases := []struct {
		name  string
		input UpdateInput
		want  error
	}{
		{"future occurrence", UpdateInput{Text: "内容", OccurredAt: now.Add(time.Second)}, ErrInvalidOccurredAt},
		{"before product epoch", UpdateInput{Text: "内容", OccurredAt: time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC)}, ErrInvalidOccurredAt},
		{"latitude without longitude", UpdateInput{Text: "内容", OccurredAt: now, Latitude: &validLatitude}, ErrInvalidCoordinates},
		{"latitude out of range", UpdateInput{Text: "内容", OccurredAt: now, Latitude: &invalidLatitude, Longitude: &longitude}, ErrInvalidCoordinates},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if _, err := moment.Update(7, item.input, now); !errors.Is(err, item.want) {
				t.Fatalf("err=%v want=%v", err, item.want)
			}
		})
	}
}
