package together

import "errors"

var (
	ErrInvalidTitle       = errors.New("invalid title")
	ErrInvalidDate        = errors.New("invalid date")
	ErrInvalidDescription = errors.New("invalid description")
	ErrNotCreator         = errors.New("not creator")
	ErrInvalidMoment      = errors.New("invalid linked moment")
	ErrNotFound           = errors.New("together item not found")
)
