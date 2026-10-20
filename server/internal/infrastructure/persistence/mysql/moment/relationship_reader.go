package infrastructure

import (
	"context"
	application "github.com/nanfang/just-for-you/server/internal/application/moment"
	relationshipapplication "github.com/nanfang/just-for-you/server/internal/application/relationship"
)

type RelationshipService interface {
	Execute(context.Context, relationshipapplication.GetRelationshipQuery) (relationshipapplication.CurrentRelationship, error)
}
type ActiveRelationshipReader struct{ service RelationshipService }

func NewActiveRelationshipReader(service RelationshipService) ActiveRelationshipReader {
	return ActiveRelationshipReader{service: service}
}
func (r ActiveRelationshipReader) GetActiveRelationship(ctx context.Context, userID int64) (application.ActiveRelationship, error) {
	current, err := r.service.Execute(ctx, relationshipapplication.GetRelationshipQuery{UserID: userID})
	if err != nil {
		return application.ActiveRelationship{}, err
	}
	relationship := current.Relationship
	partner := relationship.UserAID()
	if partner == userID {
		partner = relationship.UserBID()
	}
	return application.ActiveRelationship{ID: relationship.ID(), PartnerID: partner}, nil
}
