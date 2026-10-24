package domain

import (
	"errors"
	"strings"
	"time"
)

type Purpose string

const (
	PurposeAvatar           Purpose = "avatar"
	PurposeMomentPhoto      Purpose = "moment_photo"
	PurposeAnniversaryCover Purpose = "anniversary_cover"
)

var (
	ErrInvalidGrant  = errors.New("invalid upload grant")
	ErrGrantConsumed = errors.New("upload grant already consumed")
)

type GrantInput struct {
	ObjectKey           string
	OwnerID, ResourceID int64
	Purpose             Purpose
	ContentType         string
	Size                int64
	ExpiresAt           time.Time
}
type Grant struct {
	input      GrantInput
	consumedAt *time.Time
}

func NewGrant(in GrantInput, now time.Time) (Grant, error) {
	in.ObjectKey = strings.TrimSpace(in.ObjectKey)
	in.ContentType = strings.TrimSpace(in.ContentType)
	max := int64(5 << 20)
	if in.Purpose == PurposeMomentPhoto {
		max = 10 << 20
	}
	validPurpose := in.Purpose == PurposeAvatar || in.Purpose == PurposeMomentPhoto || in.Purpose == PurposeAnniversaryCover
	validResource := (in.Purpose == PurposeAnniversaryCover && in.ResourceID > 0) || (in.Purpose != PurposeAnniversaryCover && in.ResourceID == 0)
	validType := in.ContentType == "image/jpeg" || in.ContentType == "image/png" || in.ContentType == "image/webp"
	if in.ObjectKey == "" || in.OwnerID < 1 || !validPurpose || !validResource || !validType || in.Size < 1 || in.Size > max || !in.ExpiresAt.After(now) {
		return Grant{}, ErrInvalidGrant
	}
	return Grant{input: in}, nil
}
func RestoreGrant(in GrantInput, consumedAt *time.Time) Grant {
	return Grant{input: in, consumedAt: consumedAt}
}
func (g Grant) ObjectKey() string      { return g.input.ObjectKey }
func (g Grant) OwnerID() int64         { return g.input.OwnerID }
func (g Grant) ResourceID() int64      { return g.input.ResourceID }
func (g Grant) Purpose() Purpose       { return g.input.Purpose }
func (g Grant) ContentType() string    { return g.input.ContentType }
func (g Grant) Size() int64            { return g.input.Size }
func (g Grant) ExpiresAt() time.Time   { return g.input.ExpiresAt }
func (g Grant) ConsumedAt() *time.Time { return g.consumedAt }
func (g *Grant) Consume(ownerID, resourceID int64, purpose Purpose, now time.Time) error {
	if g.consumedAt != nil {
		return ErrGrantConsumed
	}
	if ownerID != g.OwnerID() || resourceID != g.ResourceID() || purpose != g.Purpose() || !g.ExpiresAt().After(now) {
		return ErrInvalidGrant
	}
	v := now
	g.consumedAt = &v
	return nil
}
