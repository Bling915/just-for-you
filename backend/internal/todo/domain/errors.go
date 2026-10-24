package todo

import "errors"

var (
	ErrInvalidTitle = errors.New("invalid todo title")
	ErrNotCreator   = errors.New("not todo creator")
	ErrNotFound     = errors.New("todo not found")
)
