package dailyquestion

import (
	"strings"
	"time"
	"unicode/utf8"
)

type DailyAnswer struct {
	UserID                 int64
	Text                   string
	SubmittedAt, UpdatedAt time.Time
}
type DailyQuestion struct {
	id, coupleID                 int64
	date                         time.Time
	text, source, contextVersion string
	answers                      map[int64]DailyAnswer
	unlockedAt                   *time.Time
	createdAt, updatedAt         time.Time
}

func NewDailyQuestion(coupleID int64, date time.Time, text, source string, now time.Time) (DailyQuestion, error) {
	text = strings.TrimSpace(text)
	if coupleID <= 0 || text == "" || utf8.RuneCountInString(text) > 240 || (source != "ai" && source != "fallback") {
		return DailyQuestion{}, ErrInvalidDailyQuestion
	}
	return DailyQuestion{coupleID: coupleID, date: date, text: text, source: source, answers: map[int64]DailyAnswer{}, createdAt: now, updatedAt: now}, nil
}
func RestoreDailyQuestion(id, coupleID int64, date time.Time, text, source, contextVersion string, answers []DailyAnswer, unlockedAt *time.Time, createdAt, updatedAt time.Time) DailyQuestion {
	m := map[int64]DailyAnswer{}
	for _, a := range answers {
		m[a.UserID] = a
	}
	return DailyQuestion{id: id, coupleID: coupleID, date: date, text: text, source: source, contextVersion: contextVersion, answers: m, unlockedAt: unlockedAt, createdAt: createdAt, updatedAt: updatedAt}
}
func (q DailyQuestion) Submit(userID int64, text string, partnerID int64, now time.Time) (DailyQuestion, bool, error) {
	if q.unlockedAt != nil {
		return q, false, ErrDailyQuestionFrozen
	}
	text, err := ValidateAnswer(text)
	if err != nil || userID <= 0 || partnerID <= 0 || userID == partnerID {
		return q, false, ErrInvalidDailyAnswer
	}
	if q.answers == nil {
		q.answers = map[int64]DailyAnswer{}
	}
	q.answers[userID] = DailyAnswer{UserID: userID, Text: text, SubmittedAt: now, UpdatedAt: now}
	unlocked := false
	if _, ok := q.answers[partnerID]; ok {
		v := now
		q.unlockedAt = &v
		unlocked = true
	}
	q.updatedAt = now
	return q, unlocked, nil
}

func ValidateAnswer(value string) (string, error) {
	text := strings.TrimSpace(value)
	if text == "" || utf8.RuneCountInString(text) > 500 {
		return "", ErrInvalidDailyAnswer
	}
	return text, nil
}
func (q DailyQuestion) ID() int64              { return q.id }
func (q DailyQuestion) CoupleID() int64        { return q.coupleID }
func (q DailyQuestion) Date() time.Time        { return q.date }
func (q DailyQuestion) Text() string           { return q.text }
func (q DailyQuestion) Source() string         { return q.source }
func (q DailyQuestion) ContextVersion() string { return q.contextVersion }
func (q DailyQuestion) WithContextVersion(version string) DailyQuestion {
	q.contextVersion = strings.TrimSpace(version)
	return q
}
func (q DailyQuestion) Answers() map[int64]DailyAnswer {
	out := map[int64]DailyAnswer{}
	for k, v := range q.answers {
		out[k] = v
	}
	return out
}
func (q DailyQuestion) Unlocked() bool { return q.unlockedAt != nil }
func (q DailyQuestion) UnlockedAt() *time.Time {
	if q.unlockedAt == nil {
		return nil
	}
	v := *q.unlockedAt
	return &v
}
func (q DailyQuestion) CreatedAt() time.Time { return q.createdAt }
func (q DailyQuestion) UpdatedAt() time.Time { return q.updatedAt }
