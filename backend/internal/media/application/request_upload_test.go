package application

import (
	"context"
	"github.com/nanfang/just-for-you/server/internal/media/domain"
	"testing"
	"time"
)

type grantSaver struct{ value domain.Grant }

func (s *grantSaver) Save(_ context.Context, value domain.Grant) error { s.value = value; return nil }

type signer struct{}

func (signer) SignUpload(context.Context, string, string, time.Time) (string, error) {
	return "https://upload", nil
}

func TestRequestUploadBuildsPurposeScopedGrant(t *testing.T) {
	now := time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC)
	store := &grantSaver{}
	request := NewRequestUpload(store, signer{}, func() time.Time { return now }, func() string { return "key" })
	got, err := request.Request(context.Background(), RequestUploadCommand{OwnerID: 7, ResourceID: 9, Purpose: domain.PurposeAnniversaryCover, ContentType: "image/jpeg", Size: 100})
	if err != nil {
		t.Fatal(err)
	}
	if got.ObjectKey != "covers/7/9/key" || got.UploadURL != "https://upload" || store.value.OwnerID() != 7 || !got.ExpiresAt.Equal(now.Add(10*time.Minute)) {
		t.Fatalf("got=%+v", got)
	}
}
