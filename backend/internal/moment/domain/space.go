package domain

type SpaceType string

const (
	SpaceShared   SpaceType = "shared"
	SpacePersonal SpaceType = "personal"
)

func (s SpaceType) Valid() bool { return s == SpaceShared || s == SpacePersonal }

type Visibility string

const (
	VisibilityCouple  Visibility = "couple"
	VisibilityPrivate Visibility = "private"
	VisibilityPartner Visibility = "partner"
)

func (v Visibility) Valid() bool {
	return v == VisibilityCouple || v == VisibilityPrivate || v == VisibilityPartner
}

func validateSpace(space SpaceType, visibility Visibility, relationshipID int64) error {
	if !space.Valid() || !visibility.Valid() {
		return ErrInvalidSpaceVisibility
	}
	switch {
	case space == SpaceShared && visibility == VisibilityCouple && relationshipID > 0:
		return nil
	case space == SpacePersonal && visibility == VisibilityPrivate && relationshipID == 0:
		return nil
	case space == SpacePersonal && visibility == VisibilityPartner && relationshipID > 0:
		return nil
	default:
		return ErrInvalidSpaceVisibility
	}
}

func validateLinks(space SpaceType, sourceMomentID, dailyQuestionID *int64) error {
	if !validOptionalID(sourceMomentID) || !validOptionalID(dailyQuestionID) {
		return ErrInvalidMomentKind
	}
	if sourceMomentID != nil && space != SpaceShared {
		return ErrInvalidMomentKind
	}
	if dailyQuestionID != nil {
		if space != SpaceShared || sourceMomentID != nil {
			return ErrInvalidMomentKind
		}
	}
	return nil
}

func validOptionalID(value *int64) bool { return value == nil || *value > 0 }

func copyInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
