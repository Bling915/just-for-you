package moment

import (
	"context"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"time"
)

type ActiveRelationship struct {
	ID        int64
	PartnerID int64
}
type ActiveRelationshipReader interface {
	GetActiveRelationship(context.Context, int64) (ActiveRelationship, error)
}
type UploadInput struct {
	ContentType string
	Size        int64
}
type UploadGrant struct {
	ObjectKey   string    `json:"objectKey"`
	OwnerID     int64     `json:"-"`
	ContentType string    `json:"contentType"`
	ExpiresAt   time.Time `json:"expiresAt"`
	UploadURL   string    `json:"uploadUrl,omitempty"`
}
type UploadAuthorizer interface {
	CreateGrant(context.Context, int64, UploadInput, time.Time) (UploadGrant, error)
	ValidateGrants(context.Context, int64, []string, time.Time) ([]UploadGrant, error)
}
type PhotoStorage interface {
	SignUpload(context.Context, string, string, time.Time) (string, error)
	SignRead(context.Context, string, time.Time) (string, error)
	VerifyUploaded(context.Context, string, string, int64) error
}
type TimelinePage struct {
	Moments              []domain.Moment              `json:"moments"`
	NextCursor           string                       `json:"nextCursor,omitempty"`
	PhotoURLs            map[string]string            `json:"-"`
	Reactions            map[int64][]domain.Reaction  `json:"-"`
	Perspectives         map[int64]domain.Perspective `json:"-"`
	PerspectivePhotoURLs map[string]string            `json:"-"`
}
type CalendarDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
type MapPoint struct {
	MomentID   int64     `json:"momentId"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	OccurredAt time.Time `json:"occurredAt"`
	PlaceName  string    `json:"placeName"`
}
type MonthlyReport struct {
	Month        string `json:"month"`
	MomentCount  int    `json:"momentCount"`
	PhotoCount   int    `json:"photoCount"`
	TopPlace     string `json:"topPlace"`
	CommonStatus string `json:"commonStatus"`
}
type AnniversaryHighlight struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Date          time.Time `json:"date"`
	CountMode     string    `json:"countMode"`
	RepeatsYearly bool      `json:"repeatsYearly"`
	Pinned        bool      `json:"pinned"`
	CoverRef      string    `json:"-"`
	CoverURL      string    `json:"coverUrl"`
	Message       string    `json:"message"`
	SystemKind    string    `json:"systemKind"`
}
type AnniversaryHighlightsReader interface {
	ListPinned(context.Context, int64, int) ([]AnniversaryHighlight, error)
}
type TimelineOwner string

const (
	TimelineOwnerMe      TimelineOwner = "me"
	TimelineOwnerPartner TimelineOwner = "partner"
)

type TimelineFilter struct {
	ViewerID       int64
	RelationshipID int64
	OwnerID        int64
	SpaceType      domain.SpaceType
}

type MomentLookup struct {
	ViewerID       int64
	RelationshipID int64
	MomentID       int64
}

type MomentRepository interface {
	SaveWithGrants(context.Context, domain.Moment, []UploadGrant) (domain.Moment, error)
	UpdateWithGrants(context.Context, domain.Moment, time.Time, []UploadGrant) (domain.Moment, error)
	List(context.Context, TimelineFilter, string, int) (TimelinePage, error)
	Recent(context.Context, TimelineFilter, int) ([]domain.Moment, error)
	LatestStatusByAuthor(context.Context, int64, int64) (*domain.Moment, error)
	Find(context.Context, MomentLookup) (domain.Moment, error)
	Delete(context.Context, int64) error
	ShareToCouple(context.Context, int64, int64, int64, string, time.Time) (domain.Moment, error)
	UpsertReaction(context.Context, domain.Reaction) (domain.Reaction, error)
	DeleteReaction(context.Context, int64, int64) error
	ListReactions(context.Context, []int64) (map[int64][]domain.Reaction, error)
	SavePerspectiveWithGrants(context.Context, domain.Perspective, []UploadGrant, int64) (domain.Perspective, error)
	DeletePerspective(context.Context, int64, int64) error
	ListPerspectives(context.Context, []int64) (map[int64]domain.Perspective, error)
	ResurfaceCandidates(context.Context, int64, time.Time, int) ([]domain.Moment, error)
}
type ReadRepository interface {
	Calendar(context.Context, int64, time.Time, time.Time) ([]CalendarDay, error)
	Map(context.Context, int64, time.Time, time.Time, int) ([]MapPoint, error)
	MonthlyReport(context.Context, int64, time.Time, time.Time) (MonthlyReport, error)
}
