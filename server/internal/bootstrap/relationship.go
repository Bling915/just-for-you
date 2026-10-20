package bootstrap

import (
	"database/sql"
	relationshipapp "github.com/nanfang/just-for-you/server/internal/application/relationship"
	"github.com/nanfang/just-for-you/server/internal/infrastructure/clock"
	relationshipmysql "github.com/nanfang/just-for-you/server/internal/infrastructure/persistence/mysql/relationship"
	relationshiphttp "github.com/nanfang/just-for-you/server/internal/interfaces/http/relationship"
)

func buildRelationship(db *sql.DB) (*relationshiphttp.Handler, *relationshipapp.GetRelationshipUseCase) {
	repo := relationshipmysql.NewMySQLRepository(db)
	get := relationshipapp.NewGetRelationshipUseCase(repo)
	useCases := relationshiphttp.UseCases{CreateInvitation: relationshipapp.NewCreateInvitationUseCase(repo, clock.System{}, relationshipmysql.GenerateInvitationCode), PreviewInvitation: relationshipapp.NewPreviewInvitationUseCase(repo, clock.System{}), BindRelationship: relationshipapp.NewBindRelationshipUseCase(repo, clock.System{}), GetRelationship: get, EndRelationship: relationshipapp.NewEndRelationshipUseCase(repo, clock.System{})}
	return relationshiphttp.NewHandler(useCases), get
}
