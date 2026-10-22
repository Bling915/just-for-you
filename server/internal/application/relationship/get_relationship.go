package relationship

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"time"
)

type GetRelationshipQuery struct{ UserID int64 }
type GetRelationshipUseCase struct {
	repository Repository
	avatars    AvatarURLSigner
	clock      Clock
}

type PartnerProfile struct {
	UserID      int64
	DisplayName string
	AvatarRef   string
	AvatarURL   string
}

type CurrentRelationship struct {
	Relationship domain.Relationship
	Partner      PartnerProfile
	SoloTest     bool
}

func NewGetRelationshipUseCase(r Repository) *GetRelationshipUseCase {
	return &GetRelationshipUseCase{repository: r}
}
func NewGetRelationshipUseCaseWithAvatars(r Repository, a AvatarURLSigner, c Clock) *GetRelationshipUseCase {
	return &GetRelationshipUseCase{repository: r, avatars: a, clock: c}
}
func (u *GetRelationshipUseCase) Execute(ctx context.Context, q GetRelationshipQuery) (CurrentRelationship, error) {
	relationship, partner, soloTest, err := u.repository.GetActiveDetails(ctx, q.UserID)
	if err != nil {
		return CurrentRelationship{}, err
	}
	if partner.AvatarRef != "" && u.avatars != nil {
		partner.AvatarURL, err = u.avatars.SignRead(ctx, partner.AvatarRef, u.clock.Now().Add(15*time.Minute))
		if err != nil {
			return CurrentRelationship{}, err
		}
	} else if partner.AvatarURL == "" {
		partner.AvatarURL = partner.AvatarRef
	}
	return CurrentRelationship{Relationship: relationship, Partner: partner, SoloTest: soloTest}, nil
}
