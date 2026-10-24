package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	app "github.com/nanfang/just-for-you/server/internal/dailyquestion/application"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrGenerationUnavailable = errors.New("question generation unavailable")

type OpenAICompatible struct {
	baseURL, apiKey, model string
	client                 *http.Client
}

func NewOpenAICompatible(baseURL, apiKey, model string, client *http.Client) *OpenAICompatible {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &OpenAICompatible{strings.TrimRight(baseURL, "/"), apiKey, model, client}
}
func (c *OpenAICompatible) Generate(ctx context.Context, input app.QuestionContext) (app.GeneratedQuestion, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	prompt := "为一对情侣生成且只生成一个温柔、双方都能回答、不涉及诊断的问题。"
	if len(input.Themes) > 0 {
		prompt += "可参考这些模糊主题：" + strings.Join(input.Themes, "；")
	}
	payload := map[string]any{"model": c.model, "messages": []map[string]string{{"role": "user", "content": prompt}}, "temperature": 0.8}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return app.GeneratedQuestion{}, ErrGenerationUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(req)
	if err != nil {
		return app.GeneratedQuestion{}, ErrGenerationUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return app.GeneratedQuestion{}, fmt.Errorf("%w: provider status", ErrGenerationUnavailable)
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.NewDecoder(response.Body).Decode(&result) != nil || len(result.Choices) != 1 {
		return app.GeneratedQuestion{}, ErrGenerationUnavailable
	}
	text := strings.TrimSpace(result.Choices[0].Message.Content)
	if !validGeneratedQuestion(text) {
		return app.GeneratedQuestion{}, ErrGenerationUnavailable
	}
	return app.GeneratedQuestion{Text: text, Source: "ai", ContextVersion: "v1"}, nil
}
func validGeneratedQuestion(text string) bool {
	if text == "" || utf8.RuneCountInString(text) > 120 || strings.Count(text, "?")+strings.Count(text, "？") != 1 {
		return false
	}
	lower := strings.ToLower(text)
	for _, bad := range []string{"抑郁症", "诊断", "创伤测试", "攻击", "分手测试"} {
		if strings.Contains(lower, bad) {
			return false
		}
	}
	return true
}

type FallbackGenerator struct {
	Primary  app.QuestionGenerator
	Fallback app.QuestionGenerator
}

func (g FallbackGenerator) Generate(ctx context.Context, input app.QuestionContext) (app.GeneratedQuestion, error) {
	if g.Primary != nil {
		if value, err := g.Primary.Generate(ctx, input); err == nil {
			return value, nil
		}
	}
	return g.Fallback.Generate(ctx, input)
}
