package anniversary

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
	countMode               CountMode
	pinned                  bool
	coverRef, message       string
	systemKind              string
	createdAt, updatedAt    time.Time
}

type CountMode string

const (
	CountModeElapsed   CountMode = "elapsed"
	CountModeCountdown CountMode = "countdown"
)

type AnniversaryDraft struct {
	Title         string
	Date          time.Time
	CountMode     CountMode
	RepeatsYearly bool
	Pinned        bool
	CoverRef      string
	Message       string
}

func NewAnniversary(coupleID, creatorID int64, title string, date, now time.Time) (Anniversary, error) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > 60 {
		return Anniversary{}, ErrInvalidTitle
	}
	if date.IsZero() {
		return Anniversary{}, ErrInvalidDate
	}
	return Anniversary{coupleID: coupleID, creatorID: creatorID, title: title, date: date, repeats: true, countMode: CountModeCountdown, systemKind: "none", createdAt: now, updatedAt: now}, nil
}
func RestoreAnniversary(id, coupleID, creatorID int64, title string, date time.Time, repeats bool, countMode CountMode, pinned bool, coverRef, message, systemKind string, createdAt, updatedAt time.Time) Anniversary {
	return Anniversary{id: id, coupleID: coupleID, creatorID: creatorID, title: title, date: date, repeats: repeats, countMode: countMode, pinned: pinned, coverRef: coverRef, message: message, systemKind: systemKind, createdAt: createdAt, updatedAt: updatedAt}
}
func (a Anniversary) ID() int64            { return a.id }
func (a Anniversary) CoupleID() int64      { return a.coupleID }
func (a Anniversary) CreatorID() int64     { return a.creatorID }
func (a Anniversary) Title() string        { return a.title }
func (a Anniversary) Date() time.Time      { return a.date }
func (a Anniversary) RepeatsYearly() bool  { return a.repeats }
func (a Anniversary) CountMode() CountMode { return a.countMode }
func (a Anniversary) Pinned() bool         { return a.pinned }
func (a Anniversary) CoverRef() string     { return a.coverRef }
func (a Anniversary) Message() string      { return a.message }
func (a Anniversary) SystemKind() string   { return a.systemKind }
func (a Anniversary) CreatedAt() time.Time { return a.createdAt }
func (a Anniversary) UpdatedAt() time.Time { return a.updatedAt }
func (a *Anniversary) Update(d AnniversaryDraft, now time.Time) error {
	title := strings.TrimSpace(d.Title)
	message := strings.TrimSpace(d.Message)
	coverRef := strings.TrimSpace(d.CoverRef)
	if title == "" || utf8.RuneCountInString(title) > 60 {
		return ErrInvalidTitle
	}
	if d.Date.IsZero() {
		return ErrInvalidDate
	}
	if utf8.RuneCountInString(message) > 120 {
		return ErrInvalidMessage
	}
	if d.CountMode != CountModeElapsed && d.CountMode != CountModeCountdown {
		return ErrInvalidCountMode
	}
	a.title, a.date = title, d.Date
	a.countMode, a.repeats, a.pinned = d.CountMode, d.RepeatsYearly, d.Pinned
	a.coverRef, a.message, a.updatedAt = coverRef, message, now
	return nil
}
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
