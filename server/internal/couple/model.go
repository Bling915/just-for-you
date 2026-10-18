package couple

import (
	"errors"
	"time"
)

var (
	ErrInviteInvalid  = errors.New("invite invalid")
	ErrInviteExpired  = errors.New("invite expired")
	ErrInviteUsed     = errors.New("invite used")
	ErrCannotBindSelf = errors.New("cannot bind self")
	ErrAlreadyBound   = errors.New("already bound")
	ErrNoActiveCouple = errors.New("no active couple")
)

type Invite struct {
	Code          string    `json:"code"`
	CreatorUserID int64     `json:"creatorUserId"`
	ExpiresAt     time.Time `json:"expiresAt"`
}
type InvitePreview struct {
	Code          string    `json:"code"`
	CreatorUserID int64     `json:"creatorUserId"`
	ExpiresAt     time.Time `json:"expiresAt"`
}
type Couple struct {
	ID, UserAID, UserBID int64
	Status               string
	BoundAt              time.Time
}
