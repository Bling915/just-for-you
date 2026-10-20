package domain

import "errors"

var (
	ErrInvitationInvalid    = errors.New("invitation invalid")
	ErrInvitationExpired    = errors.New("invitation expired")
	ErrInvitationUsed       = errors.New("invitation used")
	ErrCannotBindSelf       = errors.New("cannot bind self")
	ErrAlreadyBound         = errors.New("already bound")
	ErrNoActiveRelationship = errors.New("no active relationship")
)
