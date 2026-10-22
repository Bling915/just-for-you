package domain

import "errors"

var (
	ErrEmptyMoment            = errors.New("moment is empty")
	ErrTextTooLong            = errors.New("text too long")
	ErrPlaceTooLong           = errors.New("place too long")
	ErrInvalidStatus          = errors.New("invalid status")
	ErrTooManyPhotos          = errors.New("too many photos")
	ErrDuplicatePhoto         = errors.New("duplicate photo")
	ErrNotAuthor              = errors.New("not moment author")
	ErrMomentNotFound         = errors.New("moment not found")
	ErrInvalidUpload          = errors.New("invalid upload")
	ErrInvalidOccurredAt      = errors.New("invalid occurred at")
	ErrInvalidCoordinates     = errors.New("invalid coordinates")
	ErrMomentConflict         = errors.New("moment conflict")
	ErrInvalidSpaceVisibility = errors.New("invalid moment space and visibility")
	ErrInvalidMomentKind      = errors.New("invalid moment kind")
)
