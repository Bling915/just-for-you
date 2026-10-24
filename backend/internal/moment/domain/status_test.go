package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewStatusAcceptsTrimmedTextUpToTwentyRunes(t *testing.T) {
	for _, input := range []string{"", "  在上课  ", strings.Repeat("你", 20), strings.Repeat("e\u0301", 10)} {
		got, err := NewStatus(input)
		if err != nil {
			t.Fatalf("input=%q err=%v", input, err)
		}
		if input == "  在上课  " && string(got) != "在上课" {
			t.Fatalf("got=%q", got)
		}
	}
}

func TestNewStatusRejectsTwentyOneRunesAndControlCharacters(t *testing.T) {
	for _, input := range []string{strings.Repeat("你", 21), "想你\n了", "\u0000"} {
		if _, err := NewStatus(input); !errors.Is(err, ErrInvalidStatus) {
			t.Fatalf("input=%q err=%v", input, err)
		}
	}
}
