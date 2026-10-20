package bootstrap

import (
	"database/sql"
	companionshipapp "github.com/nanfang/just-for-you/server/internal/application/companionship"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/clock"
	companionshipmysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/companionship"
	companionshiphttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/companionship"
)

func buildCompanionship(db *sql.DB) *companionshiphttp.Handler {
	repo := companionshipmysql.NewRepository(db)
	return companionshiphttp.NewHandler(companionshiphttp.UseCases{UpdatePresence: companionshipapp.NewUpdatePresenceUseCase(repo, repo, clock.System{}), SendNudge: companionshipapp.NewSendNudgeUseCase(repo, repo, clock.System{}), GetSnapshot: companionshipapp.NewGetSnapshotUseCase(repo, repo, clock.System{})})
}
