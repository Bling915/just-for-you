package dailyquestionhttp

import (
	"github.com/gin-gonic/gin"
	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
)

func dailyQuestionDTO(value domain.DailyQuestion, userID int64) gin.H {
	answers := value.Answers()
	var mine any
	if answer, ok := answers[userID]; ok {
		mine = gin.H{"text": answer.Text, "submittedAt": answer.SubmittedAt}
	}
	var partner any
	if value.Unlocked() {
		for id, answer := range answers {
			if id != userID {
				partner = gin.H{"text": answer.Text, "submittedAt": answer.SubmittedAt}
			}
		}
	}
	return gin.H{
		"id":            value.ID(),
		"date":          value.Date(),
		"question":      value.Text(),
		"source":        value.Source(),
		"unlocked":      value.Unlocked(),
		"myAnswer":      mine,
		"partnerAnswer": partner,
		"answered":      mine != nil,
	}
}
