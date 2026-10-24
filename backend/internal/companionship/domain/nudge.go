package companionship

import "time"

type NudgeKind string

const (
	NudgeMissYou NudgeKind = "miss_you"
	NudgeHug     NudgeKind = "hug"
	NudgeCheerUp NudgeKind = "cheer_up"
)

type Nudge struct {
	id, coupleID, senderID, recipientID int64
	kind                                NudgeKind
	createdAt                           time.Time
}

func NewNudge(coupleID, senderID, userAID, userBID int64, kind NudgeKind, now time.Time) (Nudge, error) {
	if !kind.Valid() {
		return Nudge{}, ErrInvalidNudge
	}
	recipientID := userAID
	if senderID == userAID {
		recipientID = userBID
	} else if senderID != userBID {
		return Nudge{}, ErrNotCoupleMember
	}
	return Nudge{coupleID: coupleID, senderID: senderID, recipientID: recipientID, kind: kind, createdAt: now}, nil
}

func RestoreNudge(id, coupleID, senderID, recipientID int64, kind NudgeKind, createdAt time.Time) Nudge {
	return Nudge{id: id, coupleID: coupleID, senderID: senderID, recipientID: recipientID, kind: kind, createdAt: createdAt}
}

func (k NudgeKind) Valid() bool      { return k == NudgeMissYou || k == NudgeHug || k == NudgeCheerUp }
func (n Nudge) ID() int64            { return n.id }
func (n Nudge) CoupleID() int64      { return n.coupleID }
func (n Nudge) SenderID() int64      { return n.senderID }
func (n Nudge) RecipientID() int64   { return n.recipientID }
func (n Nudge) Kind() NudgeKind      { return n.kind }
func (n Nudge) CreatedAt() time.Time { return n.createdAt }
