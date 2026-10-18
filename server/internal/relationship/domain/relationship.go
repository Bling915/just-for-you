package domain

import "time"

type Relationship struct {
	id, userAID, userBID int64
	status               string
	boundAt              time.Time
	endedAt              *time.Time
}

func NewRelationship(id, firstID, secondID int64, boundAt time.Time) Relationship {
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}
	return Relationship{id: id, userAID: firstID, userBID: secondID, status: "active", boundAt: boundAt}
}
func RestoreRelationship(id, a, b int64, status string, boundAt time.Time, endedAt *time.Time) Relationship {
	return Relationship{id: id, userAID: a, userBID: b, status: status, boundAt: boundAt, endedAt: endedAt}
}
func (r Relationship) ID() int64           { return r.id }
func (r Relationship) UserAID() int64      { return r.userAID }
func (r Relationship) UserBID() int64      { return r.userBID }
func (r Relationship) Status() string      { return r.status }
func (r Relationship) BoundAt() time.Time  { return r.boundAt }
func (r Relationship) EndedAt() *time.Time { return r.endedAt }
