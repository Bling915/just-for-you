package companionship

import "context"

type NotificationPreferences struct {
	MomentEnabled            bool            `json:"momentEnabled"`
	PresenceEnabled          bool            `json:"presenceEnabled"`
	NudgeEnabled             bool            `json:"nudgeEnabled"`
	AIPersonalizationEnabled bool            `json:"aiPersonalizationEnabled"`
	Available                map[string]bool `json:"available"`
}
type NotificationPreferenceRepository interface {
	GetPreferences(context.Context, int64, int64) (NotificationPreferences, error)
	SavePreferences(context.Context, int64, int64, NotificationPreferences) (NotificationPreferences, error)
}
type NotificationPreferencesService struct {
	relationships RelationshipReader
	repo          NotificationPreferenceRepository
	available     map[string]bool
}

func NewNotificationPreferencesService(r RelationshipReader, repo NotificationPreferenceRepository, available map[string]bool) *NotificationPreferencesService {
	return &NotificationPreferencesService{r, repo, available}
}
func (s *NotificationPreferencesService) Get(ctx context.Context, userID int64) (NotificationPreferences, error) {
	rel, e := s.relationships.GetActiveRelationship(ctx, userID)
	if e != nil {
		return NotificationPreferences{}, e
	}
	v, e := s.repo.GetPreferences(ctx, rel.ID, userID)
	v.Available = s.available
	return v, e
}
func (s *NotificationPreferencesService) Update(ctx context.Context, userID int64, v NotificationPreferences) (NotificationPreferences, error) {
	rel, e := s.relationships.GetActiveRelationship(ctx, userID)
	if e != nil {
		return NotificationPreferences{}, e
	}
	if !s.available["moment"] {
		v.MomentEnabled = false
	}
	if !s.available["presence"] {
		v.PresenceEnabled = false
	}
	if !s.available["nudge"] {
		v.NudgeEnabled = false
	}
	v.Available = s.available
	return s.repo.SavePreferences(ctx, rel.ID, userID, v)
}
