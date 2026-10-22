package moment

import (
	"context"
	"errors"
	momentdomain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	relationshipdomain "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fakeRelationships struct {
	active ActiveRelationship
	err    error
}

func (f fakeRelationships) GetActiveRelationship(context.Context, int64) (ActiveRelationship, error) {
	return f.active, f.err
}

type fakeUploads struct {
	grants      map[string]UploadGrant
	validateErr error
}

func (f *fakeUploads) CreateGrant(_ context.Context, userID int64, input UploadInput, now time.Time) (UploadGrant, error) {
	g := UploadGrant{ObjectKey: "users/1/photo.jpg", OwnerID: userID, ContentType: input.ContentType, ExpiresAt: now.Add(10 * time.Minute)}
	f.grants[g.ObjectKey] = g
	return g, nil
}
func (f *fakeUploads) ValidateGrants(_ context.Context, userID int64, keys []string, now time.Time) ([]UploadGrant, error) {
	if f.validateErr != nil {
		return nil, f.validateErr
	}
	result := make([]UploadGrant, len(keys))
	for i, key := range keys {
		grant, ok := f.grants[key]
		if !ok || grant.OwnerID != userID || !grant.ExpiresAt.After(now) {
			return nil, momentdomain.ErrInvalidUpload
		}
		result[i] = grant
	}
	return result, nil
}

type fakeStorage struct{ verifyErr error }

func (fakeStorage) SignUpload(context.Context, string, string, time.Time) (string, error) {
	return "https://upload.test", nil
}
func (fakeStorage) SignRead(context.Context, string, time.Time) (string, error) {
	return "https://read.test", nil
}
func (f fakeStorage) VerifyUploaded(context.Context, string, string, int64) error {
	return f.verifyErr
}

type memoryMoments struct {
	saved    []momentdomain.Moment
	saveErr  error
	consumed bool
	existing momentdomain.Moment
}
type PublishInput struct {
	Text      string
	Status    momentdomain.Status
	PlaceName string
	PhotoKeys []string
}

func (r *memoryMoments) SaveWithGrants(_ context.Context, m momentdomain.Moment, _ []UploadGrant) (momentdomain.Moment, error) {
	if r.saveErr != nil {
		return momentdomain.Moment{}, r.saveErr
	}
	r.consumed = true
	r.saved = append(r.saved, m)
	return m, nil
}
func (r *memoryMoments) UpdateWithGrants(_ context.Context, m momentdomain.Moment, version time.Time, _ []UploadGrant) (momentdomain.Moment, error) {
	if !r.existing.UpdatedAt().Equal(version) {
		return momentdomain.Moment{}, momentdomain.ErrMomentConflict
	}
	r.existing = m
	return m, nil
}
func (r *memoryMoments) List(context.Context, int64, string, int) (TimelinePage, error) {
	return TimelinePage{Moments: r.saved}, nil
}
func (r *memoryMoments) Recent(context.Context, int64, int) ([]momentdomain.Moment, error) {
	return r.saved, nil
}
func (r *memoryMoments) LatestStatusByAuthor(context.Context, int64, int64) (*momentdomain.Moment, error) {
	return nil, nil
}
func (r *memoryMoments) Find(context.Context, int64, int64) (momentdomain.Moment, error) {
	return r.existing, nil
}
func (r *memoryMoments) Delete(context.Context, int64) error { return nil }

type testUseCases struct {
	publish  *PublishMomentUseCase
	timeline *GetTimelineUseCase
	delete   *DeleteMomentUseCase
}

func (u testUseCases) PublishMoment(ctx context.Context, userID int64, input PublishInput) (momentdomain.Moment, error) {
	return u.publish.Execute(ctx, PublishMomentCommand{UserID: userID, Text: input.Text, Status: input.Status, PlaceName: input.PlaceName, PhotoKeys: input.PhotoKeys})
}
func (u testUseCases) GetTimeline(ctx context.Context, userID int64, cursor string, limit int) (TimelinePage, error) {
	return u.timeline.Execute(ctx, GetTimelineQuery{userID, cursor, limit})
}
func (u testUseCases) DeleteMoment(ctx context.Context, userID, momentID int64) error {
	return u.delete.Execute(ctx, DeleteMomentCommand{userID, momentID})
}
func newMomentService(now time.Time, relationships fakeRelationships, uploads *fakeUploads, repository *memoryMoments) testUseCases {
	clock := fixedClock{now}
	storage := fakeStorage{}
	return testUseCases{NewPublishMomentUseCase(relationships, uploads, repository, storage, clock), NewGetTimelineUseCase(relationships, repository, storage, clock), NewDeleteMomentUseCase(relationships, repository)}
}
func TestPublishMomentRejectsMissingRelationship(t *testing.T) {
	service := newMomentService(time.Now(), fakeRelationships{err: relationshipdomain.ErrNoActiveRelationship}, &fakeUploads{grants: map[string]UploadGrant{}}, &memoryMoments{})
	_, err := service.PublishMoment(context.Background(), 1, PublishInput{Text: "hello"})
	if !errors.Is(err, relationshipdomain.ErrNoActiveRelationship) {
		t.Fatalf("err=%v", err)
	}
}
func TestPublishMomentRejectsInvalidPhotoGrant(t *testing.T) {
	service := newMomentService(time.Now(), fakeRelationships{active: ActiveRelationship{ID: 9}}, &fakeUploads{grants: map[string]UploadGrant{}}, &memoryMoments{})
	_, err := service.PublishMoment(context.Background(), 1, PublishInput{PhotoKeys: []string{"forged"}})
	if !errors.Is(err, momentdomain.ErrInvalidUpload) {
		t.Fatalf("err=%v", err)
	}
}

func TestPublishMomentRejectsMissingOrInvalidStoredObject(t *testing.T) {
	now := time.Now()
	uploads := &fakeUploads{grants: map[string]UploadGrant{
		"photo": {ObjectKey: "photo", OwnerID: 1, ContentType: "image/jpeg", ExpiresAt: now.Add(time.Minute)},
	}}
	service := testUseCases{publish: NewPublishMomentUseCase(fakeRelationships{active: ActiveRelationship{ID: 9}}, uploads, &memoryMoments{}, fakeStorage{verifyErr: momentdomain.ErrInvalidUpload}, fixedClock{now})}
	_, err := service.PublishMoment(context.Background(), 1, PublishInput{PhotoKeys: []string{"photo"}})
	if !errors.Is(err, momentdomain.ErrInvalidUpload) {
		t.Fatalf("err=%v", err)
	}
}

type fakeHighlights struct {
	items []AnniversaryHighlight
	err   error
	limit int
}

func (f *fakeHighlights) ListPinned(_ context.Context, _ int64, limit int) ([]AnniversaryHighlight, error) {
	f.limit = limit
	return f.items, f.err
}
func TestGetHomeLimitsHighlightsAndDegradesReaderFailure(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	moments := &memoryMoments{saved: []momentdomain.Moment{}}
	// 恰好等于展示上限：不触发「还有更多」。
	reader := &fakeHighlights{items: []AnniversaryHighlight{{ID: 1}, {ID: 2}, {ID: 3}}}
	u := NewGetHomeUseCase(fakeRelationships{active: ActiveRelationship{ID: 9, PartnerID: 2}}, moments, fakeStorage{}, fixedClock{now}, reader)
	got, err := u.Execute(context.Background(), GetHomeQuery{UserID: 1})
	if err != nil || reader.limit != pinnedDisplayLimit+1 || len(got.PinnedAnniversaries) != pinnedDisplayLimit || got.HasMorePinned {
		t.Fatalf("正好上限时 got=%+v limit=%d err=%v", got, reader.limit, err)
	}
	// 超出上限：只展示前 3 个，并置 HasMorePinned 告知前端。
	reader = &fakeHighlights{items: []AnniversaryHighlight{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}}
	u = NewGetHomeUseCase(fakeRelationships{active: ActiveRelationship{ID: 9, PartnerID: 2}}, moments, fakeStorage{}, fixedClock{now}, reader)
	got, err = u.Execute(context.Background(), GetHomeQuery{UserID: 1})
	if err != nil || len(got.PinnedAnniversaries) != pinnedDisplayLimit || !got.HasMorePinned {
		t.Fatalf("超出上限时 got=%+v err=%v", got, err)
	}
	// 读取器故障：降级为空，不影响其余数据。
	reader = &fakeHighlights{err: errors.New("anniversary unavailable")}
	u = NewGetHomeUseCase(fakeRelationships{active: ActiveRelationship{ID: 9, PartnerID: 2}}, moments, fakeStorage{}, fixedClock{now}, reader)
	got, err = u.Execute(context.Background(), GetHomeQuery{UserID: 1})
	if err != nil || len(got.PinnedAnniversaries) != 0 || got.Recent == nil {
		t.Fatalf("读取降级 got=%+v err=%v", got, err)
	}
}
func TestPublishMomentKeepsNinePhotoOrder(t *testing.T) {
	now := time.Now()
	uploads := &fakeUploads{grants: map[string]UploadGrant{}}
	keys := make([]string, 9)
	for i := range keys {
		keys[i] = string(rune('a' + i))
		uploads.grants[keys[i]] = UploadGrant{ObjectKey: keys[i], OwnerID: 1, ExpiresAt: now.Add(time.Minute)}
	}
	repository := &memoryMoments{}
	service := newMomentService(now, fakeRelationships{active: ActiveRelationship{ID: 9}}, uploads, repository)
	moment, err := service.PublishMoment(context.Background(), 1, PublishInput{PhotoKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	for i, photo := range moment.Photos() {
		if photo.ObjectKey() != keys[i] {
			t.Fatalf("photo %d=%q", i, photo.ObjectKey())
		}
	}
}
func TestPublishFailureDoesNotConsumeGrants(t *testing.T) {
	now := time.Now()
	uploads := &fakeUploads{grants: map[string]UploadGrant{"a": {ObjectKey: "a", OwnerID: 1, ExpiresAt: now.Add(time.Minute)}}}
	repository := &memoryMoments{saveErr: errors.New("database down")}
	service := newMomentService(now, fakeRelationships{active: ActiveRelationship{ID: 9}}, uploads, repository)
	_, _ = service.PublishMoment(context.Background(), 1, PublishInput{PhotoKeys: []string{"a"}})
	if repository.consumed {
		t.Fatal("grant consumed after failed save")
	}
}
func TestDeleteMomentRejectsNonAuthorAndEndedRelationship(t *testing.T) {
	now := time.Now()
	existing, _ := momentdomain.NewMoment(2, 9, "hello", momentdomain.StatusNone, "", nil, now)
	repository := &memoryMoments{existing: existing}
	service := newMomentService(now, fakeRelationships{active: ActiveRelationship{ID: 9}}, &fakeUploads{grants: map[string]UploadGrant{}}, repository)
	if err := service.DeleteMoment(context.Background(), 1, 7); !errors.Is(err, momentdomain.ErrNotAuthor) {
		t.Fatalf("err=%v", err)
	}
	service = newMomentService(now, fakeRelationships{err: relationshipdomain.ErrNoActiveRelationship}, &fakeUploads{grants: map[string]UploadGrant{}}, repository)
	if err := service.DeleteMoment(context.Background(), 1, 7); !errors.Is(err, relationshipdomain.ErrNoActiveRelationship) {
		t.Fatalf("ended err=%v", err)
	}
}
func TestTimelineRejectsOldCursorAfterRelationshipEnds(t *testing.T) {
	service := newMomentService(time.Now(), fakeRelationships{err: relationshipdomain.ErrNoActiveRelationship}, &fakeUploads{grants: map[string]UploadGrant{}}, &memoryMoments{})
	if _, err := service.GetTimeline(context.Background(), 1, "old", 20); !errors.Is(err, relationshipdomain.ErrNoActiveRelationship) {
		t.Fatalf("err=%v", err)
	}
}

func TestTimelineReturnsSignedReadURLs(t *testing.T) {
	now := time.Now()
	moment, err := momentdomain.NewMoment(1, 9, "", momentdomain.StatusNone, "", []string{"users/1/photo.jpg"}, now)
	if err != nil {
		t.Fatal(err)
	}
	repository := &memoryMoments{saved: []momentdomain.Moment{moment}}
	service := newMomentService(now, fakeRelationships{active: ActiveRelationship{ID: 9}}, &fakeUploads{grants: map[string]UploadGrant{}}, repository)

	page, err := service.GetTimeline(context.Background(), 1, "", 20)
	if err != nil {
		t.Fatal(err)
	}
	if page.PhotoURLs["users/1/photo.jpg"] != "https://read.test" {
		t.Fatalf("photo URL was not signed: %#v", page.PhotoURLs)
	}
}

func TestUpdateMomentChecksRelationshipAuthorAndNewPhotoGrant(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	existing := momentdomain.RestoreMomentDetails(5, 7, 9, "旧内容", momentdomain.StatusNone, "", nil, nil, []momentdomain.Photo{momentdomain.NewPhoto("existing", 1)}, now.Add(-time.Hour), now.Add(-time.Hour), now.Add(-time.Hour), nil)
	repository := &memoryMoments{existing: existing}
	uploads := &fakeUploads{grants: map[string]UploadGrant{"new": {ObjectKey: "new", OwnerID: 7, ContentType: "image/jpeg", ExpiresAt: now.Add(time.Hour)}}}
	useCase := NewUpdateMomentUseCase(fakeRelationships{active: ActiveRelationship{ID: 9}}, uploads, repository, fakeStorage{}, fixedClock{now})

	updated, err := useCase.Execute(context.Background(), UpdateMomentCommand{UserID: 7, MomentID: 5, Text: "新内容", PhotoKeys: []string{"new", "existing"}, OccurredAt: now.Add(-time.Minute), Version: existing.UpdatedAt()})
	if err != nil || updated.Text() != "新内容" || updated.Photos()[0].ObjectKey() != "new" {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if _, err = useCase.Execute(context.Background(), UpdateMomentCommand{UserID: 8, MomentID: 5, Text: "越权", OccurredAt: now, Version: existing.UpdatedAt()}); !errors.Is(err, momentdomain.ErrNotAuthor) {
		t.Fatalf("non-author err=%v", err)
	}
	ended := NewUpdateMomentUseCase(fakeRelationships{err: relationshipdomain.ErrNoActiveRelationship}, uploads, repository, fakeStorage{}, fixedClock{now})
	if _, err = ended.Execute(context.Background(), UpdateMomentCommand{UserID: 7, MomentID: 5, Text: "内容", OccurredAt: now, Version: existing.UpdatedAt()}); !errors.Is(err, relationshipdomain.ErrNoActiveRelationship) {
		t.Fatalf("ended err=%v", err)
	}
}
