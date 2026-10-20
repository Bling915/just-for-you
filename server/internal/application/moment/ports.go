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
	Moments    []domain.Moment   `json:"moments"`
	NextCursor string            `json:"nextCursor,omitempty"`
	PhotoURLs  map[string]string `json:"-"`
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
type MomentRepository interface {
	SaveWithGrants(context.Context, domain.Moment, []UploadGrant) (domain.Moment, error)
	UpdateWithGrants(context.Context, domain.Moment, time.Time, []UploadGrant) (domain.Moment, error)
	List(context.Context, int64, string, int) (TimelinePage, error)
	Recent(context.Context, int64, int) ([]domain.Moment, error)
	LatestStatusByAuthor(context.Context, int64, int64) (*domain.Moment, error)
	Find(context.Context, int64, int64) (domain.Moment, error)
	Delete(context.Context, int64) error
}
type ReadRepository interface {
	Calendar(context.Context, int64, time.Time, time.Time) ([]CalendarDay, error)
	Map(context.Context, int64, time.Time, time.Time, int) ([]MapPoint, error)
	MonthlyReport(context.Context, int64, time.Time, time.Time) (MonthlyReport, error)
}
