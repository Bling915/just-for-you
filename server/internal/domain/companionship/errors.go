package companionship

import "errors"

var (
	ErrInvalidPresence       = errors.New("invalid presence")
	ErrPresenceTextTooLong   = errors.New("presence text too long")
	ErrInvalidPresenceExpiry = errors.New("invalid presence expiry")
	ErrInvalidNudge          = errors.New("invalid nudge")
	ErrNotRelationshipMember = errors.New("not relationship member")
	ErrNudgeThrottled        = errors.New("nudge throttled")
)
