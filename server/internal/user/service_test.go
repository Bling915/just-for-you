package user

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateUpdateTrimsAndAcceptsThirtyRunes(t *testing.T) {
	got, e := ValidateUpdate(UpdateInput{DisplayName: "  这是一个名字  ", AvatarRef: "https://example.test/a.png"})
	if e != nil || got.DisplayName != "这是一个名字" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
func TestValidateUpdateRejectsEmptyOrOverThirtyRunes(t *testing.T) {
	for _, name := range []string{"   ", strings.Repeat("界", 31)} {
		if _, e := ValidateUpdate(UpdateInput{DisplayName: name}); !errors.Is(e, ErrInvalidProfile) {
			t.Fatalf("name=%q err=%v", name, e)
		}
	}
}
