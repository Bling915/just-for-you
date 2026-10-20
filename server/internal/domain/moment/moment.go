package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

type Moment struct {
	id, authorID, relationshipID int64
	text                         string
	status                       Status
	place                        string
	photos                       []Photo
	createdAt                    time.Time
	deletedAt                    *time.Time
}

func NewMoment(authorID, relationshipID int64, text string, status Status, place string, photoKeys []string, createdAt time.Time) (Moment, error) {
	text = strings.TrimSpace(text)
	place = strings.TrimSpace(place)
	if utf8.RuneCountInString(text) > 500 {
		return Moment{}, ErrTextTooLong
	}
	if utf8.RuneCountInString(place) > 80 {
		return Moment{}, ErrPlaceTooLong
	}
	if !status.Valid() {
		return Moment{}, ErrInvalidStatus
	}
	if len(photoKeys) > 9 {
		return Moment{}, ErrTooManyPhotos
	}
	seen := map[string]struct{}{}
	photos := make([]Photo, len(photoKeys))
	for index, key := range photoKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			return Moment{}, ErrInvalidUpload
		}
		if _, ok := seen[key]; ok {
			return Moment{}, ErrDuplicatePhoto
		}
		seen[key] = struct{}{}
		photos[index] = NewPhoto(key, index+1)
	}
	if text == "" && status == StatusNone && place == "" && len(photos) == 0 {
		return Moment{}, ErrEmptyMoment
	}
	return Moment{authorID: authorID, relationshipID: relationshipID, text: text, status: status, place: place, photos: photos, createdAt: createdAt}, nil
}
func RestoreMoment(id, authorID, relationshipID int64, text string, status Status, place string, photos []Photo, createdAt time.Time, deletedAt *time.Time) Moment {
	return Moment{id: id, authorID: authorID, relationshipID: relationshipID, text: text, status: status, place: place, photos: append([]Photo(nil), photos...), createdAt: createdAt, deletedAt: deletedAt}
}
func (m Moment) ID() int64             { return m.id }
func (m Moment) AuthorID() int64       { return m.authorID }
func (m Moment) RelationshipID() int64 { return m.relationshipID }
func (m Moment) Text() string          { return m.text }
func (m Moment) Status() Status        { return m.status }
func (m Moment) Place() string         { return m.place }
func (m Moment) Photos() []Photo       { return append([]Photo(nil), m.photos...) }
func (m Moment) CreatedAt() time.Time  { return m.createdAt }
func (m Moment) DeletedAt() *time.Time { return m.deletedAt }
