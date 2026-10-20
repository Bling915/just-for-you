package domain

import (
	"strings"
	"time"
)

const invitationAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

type Invitation struct {
	code       string
	creatorID  int64
	expiresAt  time.Time
	usedAt     *time.Time
	consumerID int64
}

func NewInvitation(code string, creatorID int64, expiresAt time.Time) (Invitation, error) {
	if len(code) != 6 {
		return Invitation{}, ErrInvitationInvalid
	}
	for _, character := range code {
		if !strings.ContainsRune(invitationAlphabet, character) {
			return Invitation{}, ErrInvitationInvalid
		}
	}
	return Invitation{code: code, creatorID: creatorID, expiresAt: expiresAt}, nil
}
func RestoreInvitation(code string, creatorID int64, expiresAt time.Time, usedAt *time.Time, consumerID int64) Invitation {
	return Invitation{code: code, creatorID: creatorID, expiresAt: expiresAt, usedAt: usedAt, consumerID: consumerID}
}
func (i Invitation) CanBeAcceptedBy(userID int64, now time.Time) error {
	if i.usedAt != nil {
		return ErrInvitationUsed
	}
	if !i.expiresAt.After(now) {
		return ErrInvitationExpired
	}
	if userID == i.creatorID {
		return ErrCannotBindSelf
	}
	return nil
}
func (i *Invitation) MarkUsed(now time.Time, consumerID int64) {
	i.usedAt = &now
	i.consumerID = consumerID
}
func (i Invitation) Code() string         { return i.code }
func (i Invitation) CreatorID() int64     { return i.creatorID }
func (i Invitation) ExpiresAt() time.Time { return i.expiresAt }
func (i Invitation) UsedAt() *time.Time   { return i.usedAt }
func (i Invitation) ConsumerID() int64    { return i.consumerID }
