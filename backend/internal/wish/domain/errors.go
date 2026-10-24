package wish

import "errors"

var (
	ErrInvalidTitle       = errors.New("invalid wish title")
	ErrInvalidDescription = errors.New("invalid wish description")
	ErrNotCreator         = errors.New("not wish creator")
	ErrInvalidMoment      = errors.New("invalid linked moment")
	ErrNotFound           = errors.New("wish not found")
)
