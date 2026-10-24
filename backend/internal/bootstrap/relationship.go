package bootstrap

import (
	"database/sql"
	relationshipapp "github.com/nanfang/just-for-you/server/internal/application/relationship"
	relationshipmysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/relationship"
	relationshiphttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/relationship"
	"github.com/nanfang/just-for-you/server/internal/platform/clock"
)

func buildRelationship(db *sql.DB, enableSoloTest bool, avatars relationshipapp.AvatarURLSigner) (*relationshiphttp.Handler, *relationshipapp.GetRelationshipUseCase) {
	repo := relationshipmysql.NewMySQLRepository(db)
	get := relationshipapp.NewGetRelationshipUseCaseWithAvatars(repo, avatars, clock.System{})
	useCases := relationshiphttp.UseCases{CreateInvitation: relationshipapp.NewCreateInvitationUseCase(repo, clock.System{}, relationshipmysql.GenerateInvitationCode), PreviewInvitation: relationshipapp.NewPreviewInvitationUseCase(repo, clock.System{}), BindRelationship: relationshipapp.NewBindRelationshipUseCase(repo, clock.System{}), GetRelationship: get, EndRelationship: relationshipapp.NewEndRelationshipUseCase(repo, clock.System{})}
	if enableSoloTest {
		useCases.EnableSoloTest = relationshipapp.NewEnableSoloTestUseCase(repo, clock.System{})
	}
	return relationshiphttp.NewHandler(useCases), get
}
