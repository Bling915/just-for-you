package bootstrap

import (
	"database/sql"
	coupleapp "github.com/nanfang/just-for-you/server/internal/couple/application"
	couplemysql "github.com/nanfang/just-for-you/server/internal/couple/infrastructure/mysql"
	couplehttp "github.com/nanfang/just-for-you/server/internal/couple/interfaces/http"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
)

func buildCouple(db *sql.DB, enableSoloTest bool, avatars coupleapp.AvatarURLSigner) (*couplehttp.Handler, *coupleapp.GetCurrentCouple) {
	repo := couplemysql.NewMySQLCoupleStore(db)
	get := coupleapp.NewGetCurrentCoupleWithAvatars(repo, avatars, clock.System{})
	useCases := couplehttp.Dependencies{CreateInvitation: coupleapp.NewCreateInvitation(repo, clock.System{}, couplemysql.GenerateInvitationCode), PreviewInvitation: coupleapp.NewPreviewInvitation(repo, clock.System{}), BindCouple: coupleapp.NewBindCouple(repo, clock.System{}), GetCouple: get, EndCouple: coupleapp.NewEndCouple(repo, clock.System{})}
	if enableSoloTest {
		useCases.EnableSoloTest = coupleapp.NewEnableSoloTest(repo, clock.System{})
	}
	return couplehttp.NewHandler(useCases), get
}
