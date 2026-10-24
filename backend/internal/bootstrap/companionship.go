package bootstrap

import (
	"database/sql"

	companionshipapp "github.com/nanfang/just-for-you/server/internal/companionship/application"
	companionshipmysql "github.com/nanfang/just-for-you/server/internal/companionship/infrastructure/mysql"
	companionshiphttp "github.com/nanfang/just-for-you/server/internal/companionship/interfaces/http"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/platform/config"
)

func buildCompanionship(
	db *sql.DB,
	cfg config.Config,
	getCouple *coupleapp.GetCurrentCouple,
) *companionshiphttp.Handler {
	relationships := newCompanionshipCoupleReader(getCouple)
	presences := companionshipmysql.NewMySQLPresenceStore(db)
	nudges := companionshipmysql.NewMySQLNudgeStore(db)
	preferences := companionshipmysql.NewMySQLNotificationPreferencesStore(db)
	available := map[string]bool{
		"moment":   cfg.WeChatMomentTemplateID != "",
		"presence": cfg.WeChatPresenceTemplateID != "",
		"nudge":    cfg.WeChatNudgeTemplateID != "",
	}
	return companionshiphttp.NewHandler(companionshiphttp.Dependencies{
		UpdatePresence:                companionshipapp.NewUpdatePresence(relationships, presences, clock.System{}),
		SendNudge:                     companionshipapp.NewSendNudge(relationships, nudges, clock.System{}),
		GetSnapshot:                   companionshipapp.NewGetSnapshot(relationships, presences, nudges, clock.System{}),
		GetNotificationPreferences:    companionshipapp.NewGetNotificationPreferences(relationships, preferences, available),
		UpdateNotificationPreferences: companionshipapp.NewUpdateNotificationPreferences(relationships, preferences, available),
	})
}
