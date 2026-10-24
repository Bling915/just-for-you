package dailyquestion

import "errors"

var (
	ErrNotFound             = errors.New("together item not found")
	ErrInvalidDailyQuestion = errors.New("invalid daily question")
	ErrInvalidDailyAnswer   = errors.New("invalid daily answer")
	ErrDailyQuestionFrozen  = errors.New("daily question is frozen")
)
