package infrastructure

import (
	"context"
	relationshipapplication "github.com/nanfang/just-for-you/server/internal/application/relationship"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"github.com/nanfang/just-for-you/server/internal/moment/application"
)

type RelationshipService interface {
	Execute(context.Context, relationshipapplication.GetRelationshipQuery) (relationshipdomain.Relationship, error)
}
type ActiveRelationshipReader struct{ service RelationshipService }

func NewActiveRelationshipReader(service RelationshipService) ActiveRelationshipReader {
	return ActiveRelationshipReader{service: service}
}
func (r ActiveRelationshipReader) GetActiveRelationship(ctx context.Context, userID int64) (application.ActiveRelationship, error) {
	relationship, err := r.service.Execute(ctx, relationshipapplication.GetRelationshipQuery{UserID: userID})
	if err != nil {
		return application.ActiveRelationship{}, err
	}
	partner := relationship.UserAID()
	if partner == userID {
		partner = relationship.UserBID()
	}
	return application.ActiveRelationship{ID: relationship.ID(), PartnerID: partner}, nil
}
