package domain

import "time"

type Couple struct {
	id, userAID, userBID int64
	status               string
	boundAt              time.Time
	endedAt              *time.Time
}

func NewCouple(id, firstID, secondID int64, boundAt time.Time) Couple {
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}
	return Couple{id: id, userAID: firstID, userBID: secondID, status: "active", boundAt: boundAt}
}
func RestoreCouple(id, a, b int64, status string, boundAt time.Time, endedAt *time.Time) Couple {
	return Couple{id: id, userAID: a, userBID: b, status: status, boundAt: boundAt, endedAt: endedAt}
}
func (r Couple) ID() int64           { return r.id }
func (r Couple) UserAID() int64      { return r.userAID }
func (r Couple) UserBID() int64      { return r.userBID }
func (r Couple) Status() string      { return r.status }
func (r Couple) BoundAt() time.Time  { return r.boundAt }
func (r Couple) EndedAt() *time.Time { return r.endedAt }
func (r Couple) PartnerOf(userID int64) (int64, error) {
	if userID == r.userAID {
		return r.userBID, nil
	}
	if userID == r.userBID {
		return r.userAID, nil
	}
	return 0, ErrNotCoupleMember
}
