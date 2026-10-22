package together

import (
	"strings"
	"time"
)

type ContextMoment struct {
	Text, Status, Place, SpaceType, Visibility string
	OccurredAt                                 time.Time
	Latitude, Longitude                        *float64
	PhotoKeys                                  []string
}
type QuestionContextBuilder struct{}

func (QuestionContextBuilder) Build(coupleID int64, date time.Time, values []ContextMoment) QuestionContext {
	themes := []string{}
	for _, value := range values {
		if value.SpaceType != "shared" || value.Visibility != "couple" {
			continue
		}
		parts := []string{}
		if text := strings.TrimSpace(value.Text); text != "" {
			r := []rune(text)
			if len(r) > 32 {
				r = r[:32]
			}
			parts = append(parts, string(r))
		}
		if value.Status != "" {
			parts = append(parts, "情绪:"+value.Status)
		}
		if value.Place != "" {
			parts = append(parts, "场景:曾一起去过的地方")
		}
		if len(parts) > 0 {
			themes = append(themes, strings.Join(parts, "；"))
		}
		if len(themes) >= 5 {
			break
		}
	}
	return QuestionContext{CoupleID: coupleID, Date: date, Themes: themes}
}
