package infrastructure

import (
	"context"
	"testing"
	"time"

	relationshipapplication "github.com/nanfang/just-for-you/server/internal/application/relationship"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
)

type currentRelationshipService struct{}

func (currentRelationshipService) Execute(context.Context, relationshipapplication.GetRelationshipQuery) (relationshipapplication.CurrentRelationship, error) {
	return relationshipapplication.CurrentRelationship{
		Relationship: relationshipdomain.NewRelationship(12, 7, 8, time.Now()),
		Partner:      relationshipapplication.PartnerProfile{UserID: 8, DisplayName: "小南"},
	}, nil
}

func TestActiveRelationshipReaderUsesStableRelationshipInsideCurrentView(t *testing.T) {
	reader := NewActiveRelationshipReader(currentRelationshipService{})

	result, err := reader.GetActiveRelationship(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 12 || result.PartnerID != 8 {
		t.Fatalf("result=%+v", result)
	}
}
