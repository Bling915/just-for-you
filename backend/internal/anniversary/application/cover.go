package anniversary

import (
	"context"
	"fmt"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type RequestCoverUploadCommand struct {
	UserID      int64
	ID          int64
	ContentType string
	Size        int64
}

type RequestCoverUpload struct {
	relationships ActiveCoupleReader
	store         AnniversaryStore
	storage       CoverStorage
	clock         Clock
	newKey        func() string
}

func NewRequestCoverUpload(
	relationships ActiveCoupleReader,
	store AnniversaryStore,
	storage CoverStorage,
	clock Clock,
	newKey func() string,
) *RequestCoverUpload {
	return &RequestCoverUpload{
		relationships: relationships,
		store:         store,
		storage:       storage,
		clock:         clock,
		newKey:        newKey,
	}
}

func (usecase *RequestCoverUpload) Execute(
	ctx context.Context,
	command RequestCoverUploadCommand,
) (CoverUploadGrant, error) {
	relationship, err := usecase.relationships.GetActiveCouple(ctx, command.UserID)
	if err != nil {
		return CoverUploadGrant{}, err
	}
	if _, err = usecase.store.Find(ctx, relationship.ID, command.ID); err != nil {
		return CoverUploadGrant{}, err
	}
	supported := command.ContentType == "image/jpeg" ||
		command.ContentType == "image/png" ||
		command.ContentType == "image/webp"
	if command.Size < 1 || command.Size > 5<<20 || !supported || usecase.storage == nil {
		return CoverUploadGrant{}, domain.ErrInvalidCover
	}
	now := usecase.clock.Now()
	grant := CoverUploadGrant{
		ObjectKey: fmt.Sprintf(
			"couples/%d/anniversaries/%d/%s",
			relationship.ID,
			command.ID,
			usecase.newKey(),
		),
		OwnerID:     command.UserID,
		ResourceID:  command.ID,
		ContentType: command.ContentType,
		Purpose:     "anniversary_cover",
		Size:        command.Size,
		ExpiresAt:   now.Add(10 * time.Minute),
	}
	if err = usecase.store.CreateCoverGrant(ctx, grant); err != nil {
		return CoverUploadGrant{}, err
	}
	grant.UploadURL, err = usecase.storage.SignUpload(
		ctx,
		grant.ObjectKey,
		grant.ContentType,
		grant.ExpiresAt,
	)
	return grant, err
}

type CoverURLQuery struct {
	CoverRef string
}

type CoverURL struct {
	storage CoverStorage
	clock   Clock
}

func NewCoverURL(storage CoverStorage, clock Clock) *CoverURL {
	return &CoverURL{storage: storage, clock: clock}
}

func (usecase *CoverURL) Execute(ctx context.Context, query CoverURLQuery) (string, error) {
	if query.CoverRef == "" || usecase.storage == nil {
		return "", nil
	}
	return usecase.storage.SignRead(
		ctx,
		query.CoverRef,
		usecase.clock.Now().Add(15*time.Minute),
	)
}
