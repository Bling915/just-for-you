package bootstrap

import (
	"database/sql"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/clock"
	mysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/together"
	togetherhttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/together"
)

func buildTogether(db *sql.DB) *togetherhttp.Handler {
	repo := mysql.NewRepository(db)
	return togetherhttp.NewHandler(app.NewService(repo, repo, clock.System{}))
}
