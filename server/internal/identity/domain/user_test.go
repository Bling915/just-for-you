package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestUserUpdateProfileTrimsNickname(t *testing.T) {
	user := NewUser(7, "openid")
	if err := user.UpdateProfile("  小明  ", "avatar/key"); err != nil {
		t.Fatal(err)
	}
	if user.DisplayName() != "小明" || user.AvatarRef() != "avatar/key" {
		t.Fatalf("profile = %q, %q", user.DisplayName(), user.AvatarRef())
	}
}

func TestUserUpdateProfileEnforcesUnicodeLength(t *testing.T) {
	for _, name := range []string{"", "   ", strings.Repeat("界", 31)} {
		user := NewUser(7, "openid")
		if err := user.UpdateProfile(name, ""); !errors.Is(err, ErrInvalidNickname) {
			t.Fatalf("name %q error = %v", name, err)
		}
	}
	user := NewUser(7, "openid")
	if err := user.UpdateProfile(strings.Repeat("界", 30), ""); err != nil {
		t.Fatalf("30-rune nickname rejected: %v", err)
	}
}
