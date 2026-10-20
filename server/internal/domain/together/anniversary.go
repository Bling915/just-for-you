package together

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Anniversary struct {
	id, coupleID, creatorID int64
	title                   string
	date                    time.Time
	repeats                 bool
	systemKind              string
	createdAt, updatedAt    time.Time
}

func NewAnniversary(coupleID, creatorID int64, title string, date, now time.Time) (Anniversary, error) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > 60 {
		return Anniversary{}, ErrInvalidTitle
	}
	if date.IsZero() {
		return Anniversary{}, ErrInvalidDate
	}
	return Anniversary{coupleID: coupleID, creatorID: creatorID, title: title, date: date, repeats: true, systemKind: "none", createdAt: now, updatedAt: now}, nil
}
func RestoreAnniversary(id, coupleID, creatorID int64, title string, date time.Time, repeats bool, systemKind string, createdAt, updatedAt time.Time) Anniversary {
	return Anniversary{id: id, coupleID: coupleID, creatorID: creatorID, title: title, date: date, repeats: repeats, systemKind: systemKind, createdAt: createdAt, updatedAt: updatedAt}
}
func (a Anniversary) ID() int64            { return a.id }
func (a Anniversary) CoupleID() int64      { return a.coupleID }
func (a Anniversary) CreatorID() int64     { return a.creatorID }
func (a Anniversary) Title() string        { return a.title }
func (a Anniversary) Date() time.Time      { return a.date }
func (a Anniversary) RepeatsYearly() bool  { return a.repeats }
func (a Anniversary) SystemKind() string   { return a.systemKind }
func (a Anniversary) CreatedAt() time.Time { return a.createdAt }
func (a Anniversary) UpdatedAt() time.Time { return a.updatedAt }
func (a Anniversary) CanDelete(userID int64) bool {
	return a.systemKind == "none" && a.creatorID == userID
}
func (a Anniversary) NextOccurrence(now time.Time) time.Time {
	if !a.repeats {
		return a.date
	}
	y := now.Year()
	month, day := a.date.Month(), a.date.Day()
	if month == time.February && day == 29 && !isLeap(y) {
		day = 28
	}
	next := time.Date(y, month, day, 0, 0, 0, 0, now.Location())
	if next.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
		y++
		day = a.date.Day()
		if month == time.February && day == 29 && !isLeap(y) {
			day = 28
		}
		next = time.Date(y, month, day, 0, 0, 0, 0, now.Location())
	}
	return next
}
func isLeap(y int) bool { return y%400 == 0 || (y%4 == 0 && y%100 != 0) }
