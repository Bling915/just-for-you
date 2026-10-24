package domain

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Status string

const (
	StatusNone    Status = ""
	StatusInClass Status = "在上课"
	StatusMissYou Status = "想你了"
	StatusTired   Status = "有点累"
)

func (s Status) Valid() bool {
	_, err := NewStatus(string(s))
	return err == nil
}

func NewStatus(value string) (Status, error) {
	value = strings.TrimSpace(value)
	if utf8.RuneCountInString(value) > 20 {
		return StatusNone, ErrInvalidStatus
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return StatusNone, ErrInvalidStatus
		}
	}
	return Status(value), nil
}
