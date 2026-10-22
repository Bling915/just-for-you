package together

import "errors"

var (
	ErrInvalidTitle       = errors.New("invalid title")
	ErrInvalidDate        = errors.New("invalid date")
	ErrInvalidMessage     = errors.New("invalid message")
	ErrInvalidCountMode   = errors.New("invalid count mode")
	ErrInvalidDescription = errors.New("invalid description")
	ErrNotCreator         = errors.New("not creator")
	ErrInvalidMoment      = errors.New("invalid linked moment")
	ErrNotFound           = errors.New("together item not found")
	ErrSystemAnniversary  = errors.New("system anniversary cannot be deleted")
	ErrInvalidCover       = errors.New("invalid anniversary cover")
)
