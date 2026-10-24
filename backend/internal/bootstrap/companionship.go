package bootstrap

import (
	"database/sql"
	companionshipapp "github.com/nanfang/just-for-you/server/internal/application/companionship"
	companionshipmysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/companionship"
	companionshiphttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/companionship"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
)

func buildCompanionship(db *sql.DB, cfg config.Config) *companionshiphttp.Handler {
	repo := companionshipmysql.NewRepository(db)
	h := companionshiphttp.NewHandler(companionshiphttp.UseCases{UpdatePresence: companionshipapp.NewUpdatePresenceUseCase(repo, repo, clock.System{}), SendNudge: companionshipapp.NewSendNudgeUseCase(repo, repo, clock.System{}), GetSnapshot: companionshipapp.NewGetSnapshotUseCase(repo, repo, clock.System{})})
	h.SetNotificationPreferences(companionshipapp.NewNotificationPreferencesService(repo, repo, map[string]bool{"moment": cfg.WeChatMomentTemplateID != "", "presence": cfg.WeChatPresenceTemplateID != "", "nudge": cfg.WeChatNudgeTemplateID != ""}))
	return h
}
