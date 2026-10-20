package companionship

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Preset string

const (
	PresetInClass Preset = "in_class"
	PresetMissYou Preset = "miss_you"
	PresetTired   Preset = "tired"
	PresetCustom  Preset = "custom"
)

type Presence struct {
	id, coupleID, userID int64
	preset               Preset
	customText           string
	expiresAt            *time.Time
	updatedAt            time.Time
}

func NewPresence(coupleID, userID int64, preset Preset, customText string, expiresAt *time.Time, now time.Time) (Presence, error) {
	customText = strings.TrimSpace(customText)
	if !preset.Valid() || (preset == PresetCustom && customText == "") || (preset != PresetCustom && customText != "") {
		return Presence{}, ErrInvalidPresence
	}
	if utf8.RuneCountInString(customText) > 20 {
		return Presence{}, ErrPresenceTextTooLong
	}
	if expiresAt != nil && !expiresAt.After(now) {
		return Presence{}, ErrInvalidPresenceExpiry
	}
	return Presence{coupleID: coupleID, userID: userID, preset: preset, customText: customText, expiresAt: expiresAt, updatedAt: now}, nil
}

func RestorePresence(id, coupleID, userID int64, preset Preset, customText string, expiresAt *time.Time, updatedAt time.Time) Presence {
	return Presence{id: id, coupleID: coupleID, userID: userID, preset: preset, customText: customText, expiresAt: expiresAt, updatedAt: updatedAt}
}

func (p Preset) Valid() bool {
	return p == PresetInClass || p == PresetMissYou || p == PresetTired || p == PresetCustom
}
func (p Presence) ID() int64             { return p.id }
func (p Presence) CoupleID() int64       { return p.coupleID }
func (p Presence) UserID() int64         { return p.userID }
func (p Presence) Preset() Preset        { return p.preset }
func (p Presence) CustomText() string    { return p.customText }
func (p Presence) ExpiresAt() *time.Time { return p.expiresAt }
func (p Presence) UpdatedAt() time.Time  { return p.updatedAt }
func (p Presence) Effective(now time.Time) bool {
	return p.expiresAt == nil || now.Before(*p.expiresAt)
}
