package auth

import (
	"testing"
	"time"
)

func TestTokenServiceIssuesAndParsesAccessToken(t *testing.T) {
	service := NewTokenService([]byte("01234567890123456789012345678901"))
	now := time.Now().UTC().Truncate(time.Second)
	raw, expiresAt, err := service.IssueAccess(42, now)
	if err != nil {
		t.Fatal(err)
	}
	if !expiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("expiresAt=%s", expiresAt)
	}
	userID, err := service.ParseAccess(raw)
	if err != nil || userID != 42 {
		t.Fatalf("ParseAccess()=(%d,%v)", userID, err)
	}
}

func TestTokenServiceRejectsMalformedToken(t *testing.T) {
	service := NewTokenService([]byte("01234567890123456789012345678901"))
	if _, err := service.ParseAccess("not-a-token"); err == nil {
		t.Fatal("ParseAccess() error=nil")
	}
}
