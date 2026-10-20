package relationship

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
)

type GetRelationshipQuery struct{ UserID int64 }
type GetRelationshipUseCase struct{ repository Repository }

type PartnerProfile struct {
	UserID      int64
	DisplayName string
	AvatarURL   string
}

type CurrentRelationship struct {
	Relationship domain.Relationship
	Partner      PartnerProfile
	SoloTest     bool
}

func NewGetRelationshipUseCase(r Repository) *GetRelationshipUseCase {
	return &GetRelationshipUseCase{r}
}
func (u *GetRelationshipUseCase) Execute(ctx context.Context, q GetRelationshipQuery) (CurrentRelationship, error) {
	relationship, partner, soloTest, err := u.repository.GetActiveDetails(ctx, q.UserID)
	if err != nil {
		return CurrentRelationship{}, err
	}
	return CurrentRelationship{Relationship: relationship, Partner: partner, SoloTest: soloTest}, nil
}
