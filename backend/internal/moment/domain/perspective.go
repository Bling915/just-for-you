package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Perspective struct {
	id, momentID, authorID int64
	text                   string
	photos                 []Photo
	createdAt, updatedAt   time.Time
}

func NewPerspective(momentID, authorID int64, text string, photoKeys []string, now time.Time) (Perspective, error) {
	text = strings.TrimSpace(text)
	if len(photoKeys) > 3 {
		return Perspective{}, ErrTooManyPerspectivePhotos
	}
	if utf8.RuneCountInString(text) > 500 {
		return Perspective{}, ErrPerspectiveTextTooLong
	}
	photos, err := makePhotos(photoKeys)
	if err != nil {
		return Perspective{}, err
	}
	if text == "" && len(photos) == 0 {
		return Perspective{}, ErrEmptyPerspective
	}
	return Perspective{momentID: momentID, authorID: authorID, text: text, photos: photos, createdAt: now, updatedAt: now}, nil
}

func RestorePerspective(id, momentID, authorID int64, text string, photos []Photo, createdAt, updatedAt time.Time) Perspective {
	return Perspective{id: id, momentID: momentID, authorID: authorID, text: text, photos: append([]Photo(nil), photos...), createdAt: createdAt, updatedAt: updatedAt}
}
func (p Perspective) CanEdit(userID int64) error {
	if p.authorID != userID {
		return ErrNotPerspectiveAuthor
	}
	return nil
}
func (p Perspective) ID() int64            { return p.id }
func (p Perspective) MomentID() int64      { return p.momentID }
func (p Perspective) AuthorID() int64      { return p.authorID }
func (p Perspective) Text() string         { return p.text }
func (p Perspective) Photos() []Photo      { return append([]Photo(nil), p.photos...) }
func (p Perspective) CreatedAt() time.Time { return p.createdAt }
func (p Perspective) UpdatedAt() time.Time { return p.updatedAt }
