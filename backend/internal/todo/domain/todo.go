package todo

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Todo struct {
	id, coupleID, creatorID int64
	completedBy             *int64
	title, status           string
	dueAt, completedAt      *time.Time
	createdAt, updatedAt    time.Time
}

func NewTodo(coupleID, creatorID int64, title string, dueAt *time.Time, now time.Time) (Todo, error) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > 120 {
		return Todo{}, ErrInvalidTitle
	}
	return Todo{coupleID: coupleID, creatorID: creatorID, title: title, dueAt: dueAt, status: "open", createdAt: now, updatedAt: now}, nil
}
func RestoreTodo(id, coupleID, creatorID int64, completedBy *int64, title string, dueAt *time.Time, status string, completedAt *time.Time, createdAt, updatedAt time.Time) Todo {
	return Todo{id: id, coupleID: coupleID, creatorID: creatorID, completedBy: completedBy, title: title, dueAt: dueAt, status: status, completedAt: completedAt, createdAt: createdAt, updatedAt: updatedAt}
}
func (t Todo) Complete(userID int64, now time.Time) Todo {
	if t.status == "completed" {
		return t
	}
	t.status = "completed"
	t.completedBy = &userID
	t.completedAt = &now
	t.updatedAt = now
	return t
}
func (t Todo) Reopen(now time.Time) Todo {
	t.status = "open"
	t.completedBy = nil
	t.completedAt = nil
	t.updatedAt = now
	return t
}
func (t Todo) ID() int64                   { return t.id }
func (t Todo) CoupleID() int64             { return t.coupleID }
func (t Todo) CreatorID() int64            { return t.creatorID }
func (t Todo) CompletedBy() *int64         { return t.completedBy }
func (t Todo) Title() string               { return t.title }
func (t Todo) DueAt() *time.Time           { return t.dueAt }
func (t Todo) Status() string              { return t.status }
func (t Todo) CompletedAt() *time.Time     { return t.completedAt }
func (t Todo) CreatedAt() time.Time        { return t.createdAt }
func (t Todo) UpdatedAt() time.Time        { return t.updatedAt }
func (t Todo) CanDelete(userID int64) bool { return t.creatorID == userID }
