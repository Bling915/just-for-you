package application

import (
	"errors"

	domain "github.com/nanfang/just-for-you/server/internal/moment/domain"
)

var (
	ErrMomentNotVisible = errors.New("moment not visible")
	ErrPartnerRequired  = errors.New("active partner required")
)

type AccessPolicy struct{}

func (AccessPolicy) CanRead(value domain.Moment, viewerID int64, relationship *ActiveCouple) error {
	if value.SpaceType() == domain.SpacePersonal {
		if viewerID == value.AuthorID() {
			return nil
		}
		if value.Visibility() == domain.VisibilityPartner && isCurrentPartner(value, viewerID, relationship) {
			return nil
		}
		return ErrMomentNotVisible
	}
	if value.SpaceType() == domain.SpaceShared && inOwningCouple(value, viewerID, relationship) {
		return nil
	}
	return ErrMomentNotVisible
}

func (p AccessPolicy) CanMutate(value domain.Moment, actorID int64, relationship *ActiveCouple) error {
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

func (p AccessPolicy) CanInteract(value domain.Moment, actorID int64, relationship *ActiveCouple) error {
	if actorID == value.AuthorID() {
		return ErrMomentNotVisible
	}
	if err := p.CanRead(value, actorID, relationship); err != nil {
		return err
	}
	if relationship == nil || relationship.ID != value.CoupleID() {
		return ErrPartnerRequired
	}
	return nil
}

func inOwningCouple(value domain.Moment, viewerID int64, relationship *ActiveCouple) bool {
	if relationship == nil || relationship.ID <= 0 || relationship.ID != value.CoupleID() {
		return false
	}
	return viewerID == value.AuthorID() || relationship.PartnerID == value.AuthorID()
}

func isCurrentPartner(value domain.Moment, viewerID int64, relationship *ActiveCouple) bool {
	return viewerID != value.AuthorID() && relationship != nil && relationship.ID == value.CoupleID() && relationship.PartnerID == value.AuthorID()
}
