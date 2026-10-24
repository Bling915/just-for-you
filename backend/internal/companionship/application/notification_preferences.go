package companionship

import "context"

type NotificationPreferences struct {
	MomentEnabled            bool            `json:"momentEnabled"`
	PresenceEnabled          bool            `json:"presenceEnabled"`
	NudgeEnabled             bool            `json:"nudgeEnabled"`
	AIPersonalizationEnabled bool            `json:"aiPersonalizationEnabled"`
	Available                map[string]bool `json:"available"`
}

type NotificationPreferencesStore interface {
	GetPreferences(context.Context, int64, int64) (NotificationPreferences, error)
	SavePreferences(context.Context, int64, int64, NotificationPreferences) (NotificationPreferences, error)
}

type GetNotificationPreferencesQuery struct {
	UserID int64
}

type GetNotificationPreferences struct {
	relationships ActiveCoupleReader
	store         NotificationPreferencesStore
	available     map[string]bool
}

func NewGetNotificationPreferences(
	relationships ActiveCoupleReader,
	store NotificationPreferencesStore,
	available map[string]bool,
) *GetNotificationPreferences {
	return &GetNotificationPreferences{
		relationships: relationships,
		store:         store,
		available:     available,
	}
}

func (query *GetNotificationPreferences) Execute(
	ctx context.Context,
	input GetNotificationPreferencesQuery,
) (NotificationPreferences, error) {
	relationship, err := query.relationships.GetActiveCouple(ctx, input.UserID)
	if err != nil {
		return NotificationPreferences{}, err
	}
	value, err := query.store.GetPreferences(ctx, relationship.ID, input.UserID)
	value.Available = query.available
	return value, err
}

type UpdateNotificationPreferencesCommand struct {
	UserID      int64
	Preferences NotificationPreferences
}

type UpdateNotificationPreferences struct {
	relationships ActiveCoupleReader
	store         NotificationPreferencesStore
	available     map[string]bool
}

func NewUpdateNotificationPreferences(
	relationships ActiveCoupleReader,
	store NotificationPreferencesStore,
	available map[string]bool,
) *UpdateNotificationPreferences {
	return &UpdateNotificationPreferences{
		relationships: relationships,
		store:         store,
		available:     available,
	}
}

func (command *UpdateNotificationPreferences) Execute(
	ctx context.Context,
	input UpdateNotificationPreferencesCommand,
) (NotificationPreferences, error) {
	relationship, err := command.relationships.GetActiveCouple(ctx, input.UserID)
	if err != nil {
		return NotificationPreferences{}, err
	}
	value := input.Preferences
	value.MomentEnabled = value.MomentEnabled && command.available["moment"]
	value.PresenceEnabled = value.PresenceEnabled && command.available["presence"]
	value.NudgeEnabled = value.NudgeEnabled && command.available["nudge"]
	value.Available = command.available
	return command.store.SavePreferences(ctx, relationship.ID, input.UserID, value)
}
