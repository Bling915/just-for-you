package ai

import (
	"context"
	app "github.com/nanfang/just-for-you/server/internal/dailyquestion/application"
	"hash/fnv"
)

type FallbackQuestions struct{ questions []string }

func NewFallbackQuestions() FallbackQuestions {
	return FallbackQuestions{questions: []string{"今天哪件小事让你觉得开心？", "最近有什么瞬间让你想起对方？", "如果今晚有一小时只属于你们，想怎么度过？", "今天最想和对方分享的一句话是什么？", "最近你最感谢对方做的哪件小事？", "下一次见面最想一起吃什么？"}}
}
func (f FallbackQuestions) Generate(_ context.Context, c app.QuestionContext) (app.GeneratedQuestion, error) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(c.Date.Format("2006-01-02")))
	questions := f.questions
	if len(questions) == 0 {
		questions = NewFallbackQuestions().questions
	}
	return app.GeneratedQuestion{Text: questions[int(h.Sum32())%len(questions)], Source: "fallback", ContextVersion: "v1"}, nil
}
