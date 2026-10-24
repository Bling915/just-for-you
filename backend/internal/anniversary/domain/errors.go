package anniversary

import "errors"

var (
	ErrInvalidTitle      = errors.New("invalid anniversary title")
	ErrInvalidDate       = errors.New("invalid anniversary date")
	ErrInvalidMessage    = errors.New("invalid anniversary message")
	ErrInvalidCountMode  = errors.New("invalid anniversary count mode")
	ErrNotCreator        = errors.New("not anniversary creator")
	ErrNotFound          = errors.New("anniversary not found")
	ErrSystemAnniversary = errors.New("system anniversary cannot be deleted")
	ErrInvalidCover      = errors.New("invalid anniversary cover")
)
