package together

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Wish struct {
	id, coupleID, creatorID     int64
	completedBy, linkedMomentID *int64
	title, description, status  string
	completedAt                 *time.Time
	createdAt, updatedAt        time.Time
}

func NewWish(coupleID, creatorID int64, title, description string, now time.Time) (Wish, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" || utf8.RuneCountInString(title) > 120 {
		return Wish{}, ErrInvalidTitle
	}
	if utf8.RuneCountInString(description) > 500 {
		return Wish{}, ErrInvalidDescription
	}
	return Wish{coupleID: coupleID, creatorID: creatorID, title: title, description: description, status: "planned", createdAt: now, updatedAt: now}, nil
}
func RestoreWish(id, coupleID, creatorID int64, completedBy, linkedMomentID *int64, title, description, status string, completedAt *time.Time, createdAt, updatedAt time.Time) Wish {
	return Wish{id: id, coupleID: coupleID, creatorID: creatorID, completedBy: completedBy, linkedMomentID: linkedMomentID, title: title, description: description, status: status, completedAt: completedAt, createdAt: createdAt, updatedAt: updatedAt}
}
func (w Wish) Complete(userID int64, momentID *int64, momentCoupleID int64, now time.Time) (Wish, error) {
	if momentID != nil && momentCoupleID != w.coupleID {
		return Wish{}, ErrInvalidMoment
	}
	if w.status == "completed" {
		return w, nil
	}
	w.status = "completed"
	w.completedBy = &userID
	w.linkedMomentID = momentID
	w.completedAt = &now
	w.updatedAt = now
	return w, nil
}
func (w Wish) Reopen(now time.Time) Wish {
	w.status = "planned"
	w.completedBy = nil
	w.linkedMomentID = nil
	w.completedAt = nil
	w.updatedAt = now
	return w
}
func (w Wish) ID() int64                   { return w.id }
func (w Wish) CoupleID() int64             { return w.coupleID }
func (w Wish) CreatorID() int64            { return w.creatorID }
func (w Wish) CompletedBy() *int64         { return w.completedBy }
func (w Wish) LinkedMomentID() *int64      { return w.linkedMomentID }
func (w Wish) Title() string               { return w.title }
func (w Wish) Description() string         { return w.description }
func (w Wish) Status() string              { return w.status }
func (w Wish) CompletedAt() *time.Time     { return w.completedAt }
func (w Wish) CreatedAt() time.Time        { return w.createdAt }
func (w Wish) UpdatedAt() time.Time        { return w.updatedAt }
func (w Wish) CanDelete(userID int64) bool { return w.creatorID == userID }
