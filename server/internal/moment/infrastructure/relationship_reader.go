package infrastructure

import (
	"context"
	"github.com/nanfang/just-for-you/server/internal/moment/application"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/relationship/domain"
)

type RelationshipService interface {
	GetRelationship(context.Context, int64) (relationshipdomain.Relationship, error)
}
type ActiveRelationshipReader struct{ service RelationshipService }

func NewActiveRelationshipReader(service RelationshipService) ActiveRelationshipReader {
	return ActiveRelationshipReader{service: service}
}
func (r ActiveRelationshipReader) GetActiveRelationship(ctx context.Context, userID int64) (application.ActiveRelationship, error) {
	relationship, err := r.service.GetRelationship(ctx, userID)
	if err != nil {
		return application.ActiveRelationship{}, err
	}
	partner := relationship.UserAID()
	if partner == userID {
		partner = relationship.UserBID()
	}
	return application.ActiveRelationship{ID: relationship.ID(), PartnerID: partner}, nil
}
