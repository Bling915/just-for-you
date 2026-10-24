package companionship

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewPresenceAcceptsPresetOrCustomStatus(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)

	preset, err := NewPresence(1, 7, PresetInClass, "", &expires, now)
	if err != nil || preset.Preset() != PresetInClass {
		t.Fatalf("presence=%+v err=%v", preset, err)
	}
	custom, err := NewPresence(1, 7, PresetCustom, "去图书馆啦", nil, now)
	if err != nil || custom.CustomText() != "去图书馆啦" {
		t.Fatalf("presence=%+v err=%v", custom, err)
	}
}

func TestNewPresenceRejectsInvalidInput(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Second)
	cases := []struct {
		name   string
		preset Preset
		text   string
		expiry *time.Time
		want   error
	}{
		{"unknown preset", Preset("busy"), "", nil, ErrInvalidPresence},
		{"missing custom text", PresetCustom, "", nil, ErrInvalidPresence},
		{"too long", PresetCustom, strings.Repeat("你", 21), nil, ErrPresenceTextTooLong},
		{"past expiry", PresetTired, "", &past, ErrInvalidPresenceExpiry},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			_, err := NewPresence(1, 7, item.preset, item.text, item.expiry, now)
			if !errors.Is(err, item.want) {
				t.Fatalf("err=%v want=%v", err, item.want)
			}
		})
	}
}

func TestPresenceExpiresAtConfiguredTime(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	presence, _ := NewPresence(1, 7, PresetTired, "", &expires, now)

	if !presence.Effective(expires.Add(-time.Nanosecond)) || presence.Effective(expires) {
		t.Fatal("presence expiry boundary is wrong")
	}
}
