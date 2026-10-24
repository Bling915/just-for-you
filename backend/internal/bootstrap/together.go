package bootstrap

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/ai"
	mysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/together"
	cosstorage "github.com/nanfang/just-for-you/server/internal/infrastructure/storage/cos"
	togetherhttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/together"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
	"net/http"
	"time"
)

func coverKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("cover")
	}
	return hex.EncodeToString(b)
}
func buildTogether(db *sql.DB, cfg config.Config) (*togetherhttp.Handler, error) {
	repo := mysql.NewRepository(db)
	storage, err := cosstorage.NewCOSStorage(cfg.COSBucketURL, cfg.COSSecretID, cfg.COSSecretKey)
	if err != nil {
		return nil, err
	}
	fallback := ai.NewFallbackQuestions()
	var primary app.QuestionGenerator
	if cfg.AIBaseURL != "" && cfg.AIAPIKey != "" && cfg.AIModel != "" {
		primary = ai.NewOpenAICompatible(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, &http.Client{Timeout: 3 * time.Second})
	}
	var generator app.QuestionGenerator = fallback
	if primary != nil {
		generator = app.PersonalizedQuestionGenerator{Preferences: repo, Primary: primary, Fallback: fallback}
	}
	daily := app.NewDailyQuestionService(repo, repo, generator, clock.System{}, repo)
	return togetherhttp.NewHandler(app.NewServiceWithStorage(repo, repo, clock.System{}, storage, coverKey), daily), nil
}
