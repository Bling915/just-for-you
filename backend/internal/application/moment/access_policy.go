package moment

import (
	"errors"

	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

var (
	ErrMomentNotVisible = errors.New("moment not visible")
	ErrPartnerRequired  = errors.New("active partner required")
)

type AccessPolicy struct{}

func (AccessPolicy) CanRead(value domain.Moment, viewerID int64, relationship *ActiveRelationship) error {
	if value.SpaceType() == domain.SpacePersonal {
		if viewerID == value.AuthorID() {
			return nil
		}
		if value.Visibility() == domain.VisibilityPartner && isCurrentPartner(value, viewerID, relationship) {
			return nil
		}
		return ErrMomentNotVisible
	}
	if value.SpaceType() == domain.SpaceShared && inOwningRelationship(value, viewerID, relationship) {
		return nil
	}
	return ErrMomentNotVisible
}

func (p AccessPolicy) CanMutate(value domain.Moment, actorID int64, relationship *ActiveRelationship) error {
	if actorID != value.AuthorID() {
		return domain.ErrNotAuthor
	}
	if value.SpaceType() == domain.SpacePersonal {
		return nil
	}
	if err := p.CanRead(value, actorID, relationship); err != nil {
		return ErrPartnerRequired
	}
	return nil
}

func (p AccessPolicy) CanInteract(value domain.Moment, actorID int64, relationship *ActiveRelationship) error {
	if actorID == value.AuthorID() {
		return ErrMomentNotVisible
	}
	if err := p.CanRead(value, actorID, relationship); err != nil {
		return err
	}
	if relationship == nil || relationship.ID != value.RelationshipID() {
		return ErrPartnerRequired
	}
	return nil
}

func inOwningRelationship(value domain.Moment, viewerID int64, relationship *ActiveRelationship) bool {
	if relationship == nil || relationship.ID <= 0 || relationship.ID != value.RelationshipID() {
		return false
	}
	return viewerID == value.AuthorID() || relationship.PartnerID == value.AuthorID()
}

func isCurrentPartner(value domain.Moment, viewerID int64, relationship *ActiveRelationship) bool {
	return viewerID != value.AuthorID() && relationship != nil && relationship.ID == value.RelationshipID() && relationship.PartnerID == value.AuthorID()
}
