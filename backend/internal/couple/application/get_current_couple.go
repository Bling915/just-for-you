package application

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/couple/domain"
	"time"
)

type GetCoupleQuery struct{ UserID int64 }
type GetCurrentCouple struct {
	repository CoupleStore
	avatars    AvatarURLSigner
	clock      Clock
}

type PartnerProfile struct {
	UserID      int64
	DisplayName string
	AvatarRef   string
	AvatarURL   string
}

type CurrentCouple struct {
	Couple   domain.Couple
	Partner  PartnerProfile
	SoloTest bool
}

func NewGetCurrentCouple(r CoupleStore) *GetCurrentCouple {
	return &GetCurrentCouple{repository: r}
}
func NewGetCurrentCoupleWithAvatars(r CoupleStore, a AvatarURLSigner, c Clock) *GetCurrentCouple {
	return &GetCurrentCouple{repository: r, avatars: a, clock: c}
}
func (u *GetCurrentCouple) Execute(ctx context.Context, q GetCoupleQuery) (CurrentCouple, error) {
	relationship, partner, soloTest, err := u.repository.GetActiveDetails(ctx, q.UserID)
	if err != nil {
		return CurrentCouple{}, err
	}
	if partner.AvatarRef != "" && u.avatars != nil {
		partner.AvatarURL, err = u.avatars.SignRead(ctx, partner.AvatarRef, u.clock.Now().Add(15*time.Minute))
		if err != nil {
			return CurrentCouple{}, err
		}
	} else if partner.AvatarURL == "" {
		partner.AvatarURL = partner.AvatarRef
	}
	return CurrentCouple{Couple: relationship, Partner: partner, SoloTest: soloTest}, nil
}
