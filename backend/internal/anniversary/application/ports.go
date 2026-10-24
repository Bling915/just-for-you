package anniversary

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type ActiveCouple struct {
	ID      int64
	UserAID int64
	UserBID int64
	BoundAt time.Time
}

type ActiveCoupleReader interface {
	GetActiveCouple(context.Context, int64) (ActiveCouple, error)
}

type Clock interface {
	Now() time.Time
}

type CoverUploadGrant struct {
	ObjectKey   string     `json:"objectKey"`
	UploadURL   string     `json:"uploadURL"`
	OwnerID     int64      `json:"-"`
	ResourceID  int64      `json:"-"`
	ContentType string     `json:"contentType"`
	Purpose     string     `json:"-"`
	Size        int64      `json:"-"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	ConsumedAt  *time.Time `json:"-"`
}

type AnniversaryStore interface {
	List(context.Context, int64) ([]domain.Anniversary, error)
	Save(context.Context, domain.Anniversary) (domain.Anniversary, error)
	Find(context.Context, int64, int64) (domain.Anniversary, error)
	Update(context.Context, domain.Anniversary) (domain.Anniversary, error)
	CreateCoverGrant(context.Context, CoverUploadGrant) error
	GetCoverGrant(context.Context, string) (CoverUploadGrant, error)
	UpdateWithCoverGrant(context.Context, domain.Anniversary, string, time.Time) (domain.Anniversary, error)
	Delete(context.Context, int64, int64, int64) error
}

type CoverStorage interface {
	SignUpload(context.Context, string, string, time.Time) (string, error)
	VerifyUploaded(context.Context, string, string, int64) error
	SignRead(context.Context, string, time.Time) (string, error)
}

type PinnedAnniversary struct {
	ID            int64
	Title         string
	Date          time.Time
	CountMode     string
	RepeatsYearly bool
	Pinned        bool
	CoverRef      string
	Message       string
	SystemKind    string
}

type PinnedAnniversaryReader interface {
	ListPinned(context.Context, int64, int) ([]PinnedAnniversary, error)
}
