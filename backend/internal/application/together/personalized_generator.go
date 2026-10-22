package together

import "context"

type PersonalizedQuestionGenerator struct {
	Preferences       AIPersonalizationReader
	Primary, Fallback QuestionGenerator
}

func (g PersonalizedQuestionGenerator) Generate(ctx context.Context, input QuestionContext) (GeneratedQuestion, error) {
	enabled, err := g.Preferences.AIPersonalizationEnabled(ctx, input.CoupleID)
	if err != nil || !enabled {
		return g.Fallback.Generate(ctx, input)
	}
	if g.Primary != nil {
		if value, e := g.Primary.Generate(ctx, input); e == nil {
			return value, nil
		}
	}
	return g.Fallback.Generate(ctx, input)
}
