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

type Kind string

const (
	KindStandard      Kind = "standard"
	KindDailyQuestion Kind = "daily_question"
)

func (k Kind) Valid() bool { return k == KindStandard || k == KindDailyQuestion }

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

func validateKind(kind Kind, space SpaceType, sourceMomentID, dailyQuestionID *int64) error {
	if !kind.Valid() || !validOptionalID(sourceMomentID) || !validOptionalID(dailyQuestionID) {
		return ErrInvalidMomentKind
	}
	if sourceMomentID != nil && (kind != KindStandard || space != SpaceShared) {
		return ErrInvalidMomentKind
	}
	if kind == KindDailyQuestion {
		if space != SpaceShared || dailyQuestionID == nil || sourceMomentID != nil {
			return ErrInvalidMomentKind
		}
		return nil
	}
	if dailyQuestionID != nil {
		return ErrInvalidMomentKind
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
