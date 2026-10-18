package application

import (
	"context"
	platformclock "github.com/nanfang/just-for-you/server/internal/platform/clock"
	"github.com/nanfang/just-for-you/server/internal/relationship/domain"
	"strings"
	"time"
)

type Service struct {
	repository   Repository
	clock        platformclock.Clock
	generateCode func() (string, error)
}

func NewService(repository Repository, clock platformclock.Clock, generator func() (string, error)) *Service {
	return &Service{repository: repository, clock: clock, generateCode: generator}
}
func (s *Service) CreateInvitation(ctx context.Context, userID int64) (domain.Invitation, error) {
	code, err := s.generateCode()
	if err != nil {
		return domain.Invitation{}, err
	}
	invitation, err := domain.NewInvitation(code, userID, s.clock.Now().Add(10*time.Minute))
	if err != nil {
		return domain.Invitation{}, err
	}
	return invitation, s.repository.SaveInvitation(ctx, invitation)
}
func (s *Service) PreviewInvitation(ctx context.Context, userID int64, code string) (domain.Invitation, error) {
	invitation, err := s.repository.FindInvitation(ctx, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		return domain.Invitation{}, err
	}
	if err = invitation.CanBeAcceptedBy(userID, s.clock.Now()); err != nil {
		return domain.Invitation{}, err
	}
	return invitation, nil
}
func (s *Service) BindRelationship(ctx context.Context, userID int64, code string) (domain.Relationship, error) {
	return s.repository.AcceptInvitation(ctx, strings.ToUpper(strings.TrimSpace(code)), userID, s.clock.Now())
}
func (s *Service) GetRelationship(ctx context.Context, userID int64) (domain.Relationship, error) {
	return s.repository.GetActive(ctx, userID)
}
func (s *Service) EndRelationship(ctx context.Context, userID int64) error {
	return s.repository.End(ctx, userID, s.clock.Now())
}
