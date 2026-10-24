package moment

import (
	"errors"
	"testing"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
)

func policyMoment(authorID, relationshipID int64, space domain.SpaceType, visibility domain.Visibility) domain.Moment {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	return domain.RestoreMomentDetails(domain.RestoreInput{
		ID: 7, AuthorID: authorID, CoupleID: relationshipID, Text: "小事",
		SpaceType: space, Visibility: visibility, Kind: domain.KindStandard,
		OccurredAt: now, CreatedAt: now, UpdatedAt: now,
	})
}

func TestAccessPolicyCanReadAcrossCoupleChanges(t *testing.T) {
	active := &ActiveCouple{ID: 9, PartnerID: 2}
	rebound := &ActiveCouple{ID: 10, PartnerID: 3}
	tests := []struct {
		name   string
		moment domain.Moment
		viewer int64
		rel    *ActiveCouple
		want   error
	}{
		{"shared author in active relationship", policyMoment(1, 9, domain.SpaceShared, domain.VisibilityCouple), 1, active, nil},
		{"shared current partner", policyMoment(1, 9, domain.SpaceShared, domain.VisibilityCouple), 2, &ActiveCouple{ID: 9, PartnerID: 1}, nil},
		{"shared stranger", policyMoment(1, 9, domain.SpaceShared, domain.VisibilityCouple), 3, active, ErrMomentNotVisible},
		{"shared after unbind", policyMoment(1, 9, domain.SpaceShared, domain.VisibilityCouple), 1, nil, ErrMomentNotVisible},
		{"shared after rebound", policyMoment(1, 9, domain.SpaceShared, domain.VisibilityCouple), 1, rebound, ErrMomentNotVisible},
		{"private author without relationship", policyMoment(1, 0, domain.SpacePersonal, domain.VisibilityPrivate), 1, nil, nil},
		{"private current partner", policyMoment(1, 0, domain.SpacePersonal, domain.VisibilityPrivate), 2, active, ErrMomentNotVisible},
		{"partner-visible author after unbind", policyMoment(1, 9, domain.SpacePersonal, domain.VisibilityPartner), 1, nil, nil},
		{"partner-visible current partner", policyMoment(1, 9, domain.SpacePersonal, domain.VisibilityPartner), 2, &ActiveCouple{ID: 9, PartnerID: 1}, nil},
		{"partner-visible after rebound", policyMoment(1, 9, domain.SpacePersonal, domain.VisibilityPartner), 3, rebound, ErrMomentNotVisible},
	}
	policy := AccessPolicy{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := policy.CanRead(test.moment, test.viewer, test.rel)
			if !errors.Is(err, test.want) {
				t.Fatalf("err=%v want=%v", err, test.want)
			}
		})
	}
}

func TestAccessPolicySeparatesMutationAndInteraction(t *testing.T) {
	policy := AccessPolicy{}
	shared := policyMoment(1, 9, domain.SpaceShared, domain.VisibilityCouple)
	partnerPersonal := policyMoment(1, 9, domain.SpacePersonal, domain.VisibilityPartner)
	private := policyMoment(1, 0, domain.SpacePersonal, domain.VisibilityPrivate)
	authorCouple := &ActiveCouple{ID: 9, PartnerID: 2}
	partnerCouple := &ActiveCouple{ID: 9, PartnerID: 1}

	if err := policy.CanMutate(shared, 1, authorCouple); err != nil {
		t.Fatalf("active author mutate: %v", err)
	}
	if err := policy.CanMutate(shared, 1, nil); !errors.Is(err, ErrPartnerRequired) {
		t.Fatalf("unbound shared mutate err=%v", err)
	}
	if err := policy.CanMutate(private, 1, nil); err != nil {
		t.Fatalf("private author mutate: %v", err)
	}
	if err := policy.CanMutate(partnerPersonal, 2, partnerCouple); !errors.Is(err, domain.ErrNotAuthor) {
		t.Fatalf("partner mutate err=%v", err)
	}

	if err := policy.CanInteract(shared, 2, partnerCouple); err != nil {
		t.Fatalf("partner interact shared: %v", err)
	}
	if err := policy.CanInteract(partnerPersonal, 2, partnerCouple); err != nil {
		t.Fatalf("partner interact personal: %v", err)
	}
	if err := policy.CanInteract(shared, 1, authorCouple); !errors.Is(err, ErrMomentNotVisible) {
		t.Fatalf("author interact err=%v", err)
	}
	if err := policy.CanInteract(private, 2, partnerCouple); !errors.Is(err, ErrMomentNotVisible) {
		t.Fatalf("private interact err=%v", err)
	}
}

func TestSoloModeNeverCreatesPartnerAccess(t *testing.T) {
	policy := AccessPolicy{}
	private := policyMoment(1, 0, domain.SpacePersonal, domain.VisibilityPrivate)
	solo := &ActiveCouple{ID: 9, PartnerID: 1}
	if err := policy.CanRead(private, 1, solo); err != nil {
		t.Fatalf("owner read: %v", err)
	}
	if err := policy.CanInteract(private, 1, solo); !errors.Is(err, ErrMomentNotVisible) {
		t.Fatalf("solo interaction err=%v", err)
	}
}
