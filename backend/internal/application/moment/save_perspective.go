package moment

import (
	"context"
	"errors"

	appshared "github.com/nanfang/just-for-you/server/internal/application/shared"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

var ErrPerspectiveOnlyForShared = errors.New("perspective only for shared moment")
var ErrPerspectiveRequiresPartner = errors.New("perspective must be written by the other partner")

type SavePerspectiveCommand struct {
	UserID, MomentID int64
	Text             string
	PhotoKeys        []string
}
type DeletePerspectiveCommand struct{ UserID, MomentID int64 }
type SavePerspectiveUseCase struct {
	relationships ActiveRelationshipReader
	uploads       UploadAuthorizer
	moments       MomentRepository
	photos        PhotoStorage
	clock         appshared.Clock
}

func NewSavePerspectiveUseCase(r ActiveRelationshipReader, u UploadAuthorizer, m MomentRepository, p PhotoStorage, c appshared.Clock) *SavePerspectiveUseCase {
	return &SavePerspectiveUseCase{r, u, m, p, c}
}
func (u *SavePerspectiveUseCase) Execute(ctx context.Context, command SavePerspectiveCommand) (domain.Perspective, error) {
	rel, err := u.relationships.GetActiveRelationship(ctx, command.UserID)
	if err != nil {
		return domain.Perspective{}, err
	}
	moment, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, RelationshipID: rel.ID, MomentID: command.MomentID})
	if err != nil {
		return domain.Perspective{}, err
	}
	if moment.SpaceType() != domain.SpaceShared || moment.RelationshipID() != rel.ID {
		return domain.Perspective{}, ErrPerspectiveOnlyForShared
	}
	if moment.AuthorID() == command.UserID {
		return domain.Perspective{}, ErrPerspectiveRequiresPartner
	}
	value, err := domain.NewPerspective(command.MomentID, command.UserID, command.Text, command.PhotoKeys, u.clock.Now())
	if err != nil {
		return domain.Perspective{}, err
	}
	grants, err := u.uploads.ValidateGrants(ctx, command.UserID, command.PhotoKeys, u.clock.Now())
	if err != nil {
		return domain.Perspective{}, err
	}
	for _, grant := range grants {
		if err = u.photos.VerifyUploaded(ctx, grant.ObjectKey, grant.ContentType, 10<<20); err != nil {
			return domain.Perspective{}, domain.ErrInvalidUpload
		}
	}
	return u.moments.SavePerspectiveWithGrants(ctx, value, grants, rel.ID)
}
func (u *SavePerspectiveUseCase) Delete(ctx context.Context, command DeletePerspectiveCommand) error {
	rel, err := u.relationships.GetActiveRelationship(ctx, command.UserID)
	if err != nil {
		return err
	}
	moment, err := u.moments.Find(ctx, MomentLookup{ViewerID: command.UserID, RelationshipID: rel.ID, MomentID: command.MomentID})
	if err != nil {
		return err
	}
	if moment.SpaceType() != domain.SpaceShared {
		return ErrPerspectiveOnlyForShared
	}
	return u.moments.DeletePerspective(ctx, command.MomentID, command.UserID)
}
