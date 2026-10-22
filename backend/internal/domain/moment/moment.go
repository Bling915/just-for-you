package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

var earliestOccurredAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

type Moment struct {
	id, authorID, relationshipID int64
	text                         string
	status                       Status
	spaceType                    SpaceType
	visibility                   Visibility
	kind                         Kind
	sourceMomentID               *int64
	dailyQuestionID              *int64
	place                        string
	latitude, longitude          *float64
	photos                       []Photo
	occurredAt, createdAt        time.Time
	updatedAt                    time.Time
	deletedAt                    *time.Time
}

type CreateInput struct {
	AuthorID, RelationshipID int64
	Text                     string
	Status                   Status
	Place                    string
	Latitude, Longitude      *float64
	PhotoKeys                []string
	OccurredAt, CreatedAt    time.Time
	SpaceType                SpaceType
	Visibility               Visibility
	Kind                     Kind
	SourceMomentID           *int64
	DailyQuestionID          *int64
}

type RestoreInput struct {
	ID, AuthorID, RelationshipID int64
	Text                         string
	Status                       Status
	Place                        string
	Latitude, Longitude          *float64
	Photos                       []Photo
	OccurredAt, CreatedAt        time.Time
	UpdatedAt                    time.Time
	DeletedAt                    *time.Time
	SpaceType                    SpaceType
	Visibility                   Visibility
	Kind                         Kind
	SourceMomentID               *int64
	DailyQuestionID              *int64
}

type UpdateInput struct {
	Text                string
	Status              Status
	Place               string
	Latitude, Longitude *float64
	PhotoKeys           []string
	OccurredAt          time.Time
}

func NewMoment(authorID, relationshipID int64, text string, status Status, place string, photoKeys []string, createdAt time.Time) (Moment, error) {
	return NewMomentDetails(CreateInput{AuthorID: authorID, RelationshipID: relationshipID, Text: text, Status: status, Place: place, PhotoKeys: photoKeys, OccurredAt: createdAt, CreatedAt: createdAt, SpaceType: SpaceShared, Visibility: VisibilityCouple, Kind: KindStandard})
}

func NewMomentDetails(input CreateInput) (Moment, error) {
	if err := validateSpace(input.SpaceType, input.Visibility, input.RelationshipID); err != nil {
		return Moment{}, err
	}
	if err := validateKind(input.Kind, input.SpaceType, input.SourceMomentID, input.DailyQuestionID); err != nil {
		return Moment{}, err
	}
	update := UpdateInput{Text: input.Text, Status: input.Status, Place: input.Place, Latitude: input.Latitude, Longitude: input.Longitude, PhotoKeys: input.PhotoKeys, OccurredAt: input.OccurredAt}
	fields, err := validateFields(update, input.CreatedAt)
	if err != nil {
		return Moment{}, err
	}
	return Moment{authorID: input.AuthorID, relationshipID: input.RelationshipID, text: fields.text, status: input.Status, spaceType: input.SpaceType, visibility: input.Visibility, kind: input.Kind, sourceMomentID: copyInt64(input.SourceMomentID), dailyQuestionID: copyInt64(input.DailyQuestionID), place: fields.place, latitude: copyFloat(input.Latitude), longitude: copyFloat(input.Longitude), photos: fields.photos, occurredAt: input.OccurredAt, createdAt: input.CreatedAt, updatedAt: input.CreatedAt}, nil
}

func (m Moment) Update(actorID int64, input UpdateInput, now time.Time) (Moment, error) {
	if actorID != m.authorID {
		return Moment{}, ErrNotAuthor
	}
	fields, err := validateFields(input, now)
	if err != nil {
		return Moment{}, err
	}
	m.text, m.status, m.place = fields.text, input.Status, fields.place
	m.latitude, m.longitude, m.photos = copyFloat(input.Latitude), copyFloat(input.Longitude), fields.photos
	m.occurredAt, m.updatedAt = input.OccurredAt, now
	return m, nil
}

type validatedFields struct {
	text, place string
	photos      []Photo
}

func validateFields(input UpdateInput, now time.Time) (validatedFields, error) {
	text, place := strings.TrimSpace(input.Text), strings.TrimSpace(input.Place)
	if utf8.RuneCountInString(text) > 500 {
		return validatedFields{}, ErrTextTooLong
	}
	if utf8.RuneCountInString(place) > 80 {
		return validatedFields{}, ErrPlaceTooLong
	}
	if !input.Status.Valid() {
		return validatedFields{}, ErrInvalidStatus
	}
	if input.OccurredAt.Before(earliestOccurredAt) || input.OccurredAt.After(now) {
		return validatedFields{}, ErrInvalidOccurredAt
	}
	if !validCoordinates(input.Latitude, input.Longitude) {
		return validatedFields{}, ErrInvalidCoordinates
	}
	photos, err := makePhotos(input.PhotoKeys)
	if err != nil {
		return validatedFields{}, err
	}
	if text == "" && input.Status == StatusNone && place == "" && len(photos) == 0 {
		return validatedFields{}, ErrEmptyMoment
	}
	return validatedFields{text: text, place: place, photos: photos}, nil
}
func makePhotos(keys []string) ([]Photo, error) {
	if len(keys) > 9 {
		return nil, ErrTooManyPhotos
	}
	seen := map[string]struct{}{}
	photos := make([]Photo, len(keys))
	for index, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, ErrInvalidUpload
		}
		if _, ok := seen[key]; ok {
			return nil, ErrDuplicatePhoto
		}
		seen[key] = struct{}{}
		photos[index] = NewPhoto(key, index+1)
	}
	return photos, nil
}
func validCoordinates(latitude, longitude *float64) bool {
	if (latitude == nil) != (longitude == nil) {
		return false
	}
	if latitude == nil {
		return true
	}
	return *latitude >= -90 && *latitude <= 90 && *longitude >= -180 && *longitude <= 180
}
func copyFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func RestoreMoment(id, authorID, relationshipID int64, text string, status Status, place string, photos []Photo, createdAt time.Time, deletedAt *time.Time) Moment {
	return RestoreMomentDetails(RestoreInput{ID: id, AuthorID: authorID, RelationshipID: relationshipID, Text: text, Status: status, Place: place, Photos: photos, OccurredAt: createdAt, CreatedAt: createdAt, UpdatedAt: createdAt, DeletedAt: deletedAt, SpaceType: SpaceShared, Visibility: VisibilityCouple, Kind: KindStandard})
}
func RestoreMomentDetails(input RestoreInput) Moment {
	return Moment{id: input.ID, authorID: input.AuthorID, relationshipID: input.RelationshipID, text: input.Text, status: input.Status, spaceType: input.SpaceType, visibility: input.Visibility, kind: input.Kind, sourceMomentID: copyInt64(input.SourceMomentID), dailyQuestionID: copyInt64(input.DailyQuestionID), place: input.Place, latitude: copyFloat(input.Latitude), longitude: copyFloat(input.Longitude), photos: append([]Photo(nil), input.Photos...), occurredAt: input.OccurredAt, createdAt: input.CreatedAt, updatedAt: input.UpdatedAt, deletedAt: input.DeletedAt}
}
func (m Moment) ID() int64               { return m.id }
func (m Moment) AuthorID() int64         { return m.authorID }
func (m Moment) RelationshipID() int64   { return m.relationshipID }
func (m Moment) Text() string            { return m.text }
func (m Moment) Status() Status          { return m.status }
func (m Moment) SpaceType() SpaceType    { return m.spaceType }
func (m Moment) Visibility() Visibility  { return m.visibility }
func (m Moment) Kind() Kind              { return m.kind }
func (m Moment) SourceMomentID() *int64  { return copyInt64(m.sourceMomentID) }
func (m Moment) DailyQuestionID() *int64 { return copyInt64(m.dailyQuestionID) }
func (m Moment) Place() string           { return m.place }
func (m Moment) Latitude() *float64      { return copyFloat(m.latitude) }
func (m Moment) Longitude() *float64     { return copyFloat(m.longitude) }
func (m Moment) Photos() []Photo         { return append([]Photo(nil), m.photos...) }
func (m Moment) OccurredAt() time.Time   { return m.occurredAt }
func (m Moment) CreatedAt() time.Time    { return m.createdAt }
func (m Moment) UpdatedAt() time.Time    { return m.updatedAt }
func (m Moment) DeletedAt() *time.Time   { return m.deletedAt }
