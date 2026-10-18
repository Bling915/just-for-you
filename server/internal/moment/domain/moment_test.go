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
