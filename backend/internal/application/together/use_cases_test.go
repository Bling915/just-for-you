package together

import (
	"context"
	"errors"
	relationship "github.com/nanfang/just-for-you/server/internal/couple/domain"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	"testing"
	"time"
)

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

type fakeRelations struct {
	rel ActiveCouple
	err error
}

func (f fakeRelations) GetActiveCouple(context.Context, int64) (ActiveCouple, error) {
	return f.rel, f.err
}

type fakeRepo struct {
	todo          domain.Todo
	anniversaries []domain.Anniversary
	updated       int
	coverGrants   map[string]CoverUploadGrant
}

func (f *fakeRepo) CreateCoverGrant(_ context.Context, g CoverUploadGrant) error {
	if f.coverGrants == nil {
		f.coverGrants = map[string]CoverUploadGrant{}
	}
	f.coverGrants[g.ObjectKey] = g
	return nil
}
func (f *fakeRepo) GetCoverGrant(_ context.Context, key string) (CoverUploadGrant, error) {
	g, ok := f.coverGrants[key]
	if !ok {
		return CoverUploadGrant{}, domain.ErrInvalidCover
	}
	return g, nil
}
func (f *fakeRepo) UpdateAnniversaryWithCoverGrant(_ context.Context, v domain.Anniversary, key string, now time.Time) (domain.Anniversary, error) {
	g, ok := f.coverGrants[key]
	if !ok || g.ConsumedAt != nil || g.OwnerID == 0 || g.ResourceID != v.ID() || g.Purpose != "anniversary_cover" || !g.ExpiresAt.After(now) {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	g.ConsumedAt = &now
	f.coverGrants[key] = g
	return f.UpdateAnniversary(context.Background(), v)
}

func (f *fakeRepo) ListAnniversaries(context.Context, int64) ([]domain.Anniversary, error) {
	return append([]domain.Anniversary(nil), f.anniversaries...), nil
}
func (f *fakeRepo) SaveAnniversary(_ context.Context, v domain.Anniversary) (domain.Anniversary, error) {
	if v.ID() == 0 {
		v = domain.RestoreAnniversary(100, v.CoupleID(), v.CreatorID(), v.Title(), v.Date(), v.RepeatsYearly(), v.CountMode(), v.Pinned(), v.CoverRef(), v.Message(), v.SystemKind(), v.CreatedAt(), v.UpdatedAt())
	}
	f.anniversaries = append(f.anniversaries, v)
	return v, nil
}
func (f *fakeRepo) FindAnniversary(_ context.Context, coupleID, id int64) (domain.Anniversary, error) {
	for _, item := range f.anniversaries {
		if item.CoupleID() == coupleID && item.ID() == id {
			return item, nil
		}
	}
	return domain.Anniversary{}, domain.ErrNotFound
}
func (f *fakeRepo) UpdateAnniversary(_ context.Context, v domain.Anniversary) (domain.Anniversary, error) {
	for i := range f.anniversaries {
		if f.anniversaries[i].ID() == v.ID() && f.anniversaries[i].CoupleID() == v.CoupleID() {
			f.anniversaries[i] = v
			return v, nil
		}
	}
	return domain.Anniversary{}, domain.ErrNotFound
}
func (f *fakeRepo) DeleteAnniversary(context.Context, int64, int64, int64) error { return nil }
func (f *fakeRepo) ListTodos(context.Context, int64) ([]domain.Todo, error)      { return nil, nil }
func (f *fakeRepo) SaveTodo(_ context.Context, v domain.Todo) (domain.Todo, error) {
	f.todo = v
	return v, nil
}
func (f *fakeRepo) FindTodo(context.Context, int64, int64) (domain.Todo, error) { return f.todo, nil }
func (f *fakeRepo) UpdateTodo(_ context.Context, v domain.Todo) (domain.Todo, error) {
	f.updated++
	f.todo = v
	return v, nil
}
func (f *fakeRepo) DeleteTodo(context.Context, int64, int64, int64) error          { return nil }
func (f *fakeRepo) ListWishes(context.Context, int64) ([]domain.Wish, error)       { return nil, nil }
func (f *fakeRepo) SaveWish(_ context.Context, v domain.Wish) (domain.Wish, error) { return v, nil }
func (f *fakeRepo) FindWish(context.Context, int64, int64) (domain.Wish, error) {
	return domain.Wish{}, nil
}
func (f *fakeRepo) UpdateWish(_ context.Context, v domain.Wish) (domain.Wish, error) { return v, nil }
func (f *fakeRepo) DeleteWish(context.Context, int64, int64, int64) error            { return nil }
func (f *fakeRepo) MomentCoupleID(context.Context, int64) (int64, error)             { return 1, nil }
func TestEndedCoupleDeniesTogetherData(t *testing.T) {
	s := NewService(&fakeRepo{}, fakeRelations{err: relationship.ErrNoActiveCouple}, fakeClock{})
	if _, e := s.ListTodos(context.Background(), 2); !errors.Is(e, relationship.ErrNoActiveCouple) {
		t.Fatalf("got %v", e)
	}
}
func TestTodoCompletionIsIdempotent(t *testing.T) {
	now := time.Now()
	todo, _ := domain.NewTodo(1, 2, "散步", nil, now)
	repo := &fakeRepo{todo: todo}
	s := NewService(repo, fakeRelations{rel: ActiveCouple{ID: 1, UserAID: 2, UserBID: 3}}, fakeClock{now: now})
	first, e := s.SetTodoCompleted(context.Background(), 3, 9, true)
	if e != nil || first.Status() != "completed" {
		t.Fatalf("first: %v", e)
	}
	second, e := s.SetTodoCompleted(context.Background(), 3, 9, true)
	if e != nil || second.Status() != "completed" {
		t.Fatalf("second: %v", e)
	}
	if repo.updated != 2 {
		t.Fatalf("expected persisted idempotent response")
	}
}

func TestListAnniversariesSortsPinnedCountdownThenElapsedStably(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, loc)
	item := func(id int64, date time.Time, mode domain.CountMode, pinned bool) domain.Anniversary {
		return domain.RestoreAnniversary(id, 9, 2, "纪念", date, true, mode, pinned, "", "", "none", now, now)
	}
	repo := &fakeRepo{anniversaries: []domain.Anniversary{
		item(4, now.AddDate(-1, 0, 0), domain.CountModeElapsed, false),
		item(2, now.AddDate(0, 0, 2), domain.CountModeCountdown, true),
		item(3, now.AddDate(0, 0, 1), domain.CountModeCountdown, true),
		item(1, now.AddDate(-2, 0, 0), domain.CountModeElapsed, false),
	}}
	s := NewService(repo, fakeRelations{rel: ActiveCouple{ID: 9, UserAID: 2, UserBID: 3}}, fakeClock{now})
	got, err := s.ListAnniversaries(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{3, 2, 4, 1}
	for i := range want {
		if got[i].ID() != want[i] {
			t.Fatalf("order[%d]=%d want %d", i, got[i].ID(), want[i])
		}
	}
}

func TestPartnerCanUpdateAnniversaryInsideCoupleButNotForeignItem(t *testing.T) {
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	owned := domain.RestoreAnniversary(7, 9, 2, "相识", now, true, domain.CountModeCountdown, false, "", "", "none", now, now)
	foreign := domain.RestoreAnniversary(8, 10, 4, "别人的", now, true, domain.CountModeCountdown, false, "", "", "none", now, now)
	repo := &fakeRepo{anniversaries: []domain.Anniversary{owned, foreign}}
	s := NewService(repo, fakeRelations{rel: ActiveCouple{ID: 9, UserAID: 2, UserBID: 3}}, fakeClock{now})
	draft := domain.AnniversaryDraft{Title: "我们的相识", Date: now, CountMode: domain.CountModeElapsed, Pinned: true}
	updated, err := s.UpdateAnniversary(context.Background(), 3, 7, draft)
	if err != nil || updated.Title() != "我们的相识" || !updated.Pinned() {
		t.Fatalf("updated=%#v err=%v", updated, err)
	}
	if _, err = s.UpdateAnniversary(context.Background(), 3, 8, draft); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("foreign update error=%v", err)
	}
}

func TestSystemAnniversaryCannotBeDeleted(t *testing.T) {
	now := time.Now()
	repo := &fakeRepo{anniversaries: []domain.Anniversary{domain.RestoreAnniversary(7, 9, 2, "我们在一起", now, true, domain.CountModeElapsed, true, "", "", "relationship", now, now)}}
	s := NewService(repo, fakeRelations{rel: ActiveCouple{ID: 9, UserAID: 2, UserBID: 3}}, fakeClock{now})
	if err := s.DeleteAnniversary(context.Background(), 2, 7); !errors.Is(err, domain.ErrSystemAnniversary) {
		t.Fatalf("error=%v", err)
	}
}

func TestListCreatesEditableCoupleAnniversaryOnce(t *testing.T) {
	now := time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
	repo := &fakeRepo{}
	s := NewService(repo, fakeRelations{rel: ActiveCouple{ID: 9, UserAID: 2, UserBID: 3, BoundAt: now}}, fakeClock{now})
	first, err := s.ListAnniversaries(context.Background(), 3)
	if err != nil || len(first) != 1 || first[0].ID() == 0 || first[0].SystemKind() != "relationship" {
		t.Fatalf("first=%#v err=%v", first, err)
	}
	second, err := s.ListAnniversaries(context.Background(), 3)
	if err != nil || len(second) != 1 || second[0].ID() != first[0].ID() {
		t.Fatalf("second=%#v err=%v", second, err)
	}
}

type coverStorageStub struct{ verifyErr error }

func (coverStorageStub) SignUpload(_ context.Context, key, contentType string, _ time.Time) (string, error) {
	return "https://upload/" + key, nil
}
func (c coverStorageStub) VerifyUploaded(context.Context, string, string, int64) error {
	return c.verifyErr
}
func (coverStorageStub) SignRead(_ context.Context, key string, _ time.Time) (string, error) {
	return "https://read/" + key, nil
}

func TestCoverUploadGrantRequiresOwnedAnniversaryAndSupportedImage(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	item := domain.RestoreAnniversary(7, 9, 2, "相识", now, true, domain.CountModeCountdown, false, "", "", "none", now, now)
	repo := &fakeRepo{anniversaries: []domain.Anniversary{item}}
	s := NewServiceWithStorage(repo, fakeRelations{rel: ActiveCouple{ID: 9, UserAID: 2, UserBID: 3}}, fakeClock{now}, coverStorageStub{}, func() string { return "fixed" })
	g, err := s.RequestCoverUpload(context.Background(), 3, 7, "image/jpeg", 5<<20)
	if err != nil || g.ObjectKey != "couples/9/anniversaries/7/fixed" || g.ResourceID != 7 {
		t.Fatalf("grant=%+v err=%v", g, err)
	}
	if _, err = s.RequestCoverUpload(context.Background(), 3, 8, "image/jpeg", 100); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("foreign err=%v", err)
	}
	if _, err = s.RequestCoverUpload(context.Background(), 3, 7, "image/gif", 100); !errors.Is(err, domain.ErrInvalidCover) {
		t.Fatalf("type err=%v", err)
	}
}

func TestCoverUpdateConsumesMatchingGrantOnceAndPreservesOldCoverOnFailure(t *testing.T) {
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	item := domain.RestoreAnniversary(7, 9, 2, "相识", now, true, domain.CountModeCountdown, false, "old", "", "none", now, now)
	grant := CoverUploadGrant{ObjectKey: "couples/9/anniversaries/7/new", OwnerID: 3, ResourceID: 7, ContentType: "image/jpeg", Purpose: "anniversary_cover", ExpiresAt: now.Add(time.Minute)}
	repo := &fakeRepo{anniversaries: []domain.Anniversary{item}, coverGrants: map[string]CoverUploadGrant{grant.ObjectKey: grant}}
	draft := domain.AnniversaryDraft{Title: "相识", Date: now, CountMode: domain.CountModeCountdown, RepeatsYearly: true, CoverRef: grant.ObjectKey}
	s := NewServiceWithStorage(repo, fakeRelations{rel: ActiveCouple{ID: 9, UserAID: 2, UserBID: 3}}, fakeClock{now}, coverStorageStub{verifyErr: errors.New("missing")}, func() string { return "x" })
	if _, err := s.UpdateAnniversary(context.Background(), 3, 7, draft); !errors.Is(err, domain.ErrInvalidCover) {
		t.Fatalf("verify err=%v", err)
	}
	stored, _ := repo.FindAnniversary(context.Background(), 9, 7)
	if stored.CoverRef() != "old" {
		t.Fatalf("cover changed=%s", stored.CoverRef())
	}
	s = NewServiceWithStorage(repo, fakeRelations{rel: ActiveCouple{ID: 9, UserAID: 2, UserBID: 3}}, fakeClock{now}, coverStorageStub{}, func() string { return "x" })
	if _, err := s.UpdateAnniversary(context.Background(), 3, 7, draft); err != nil {
		t.Fatal(err)
	}
	repo.anniversaries[0] = item
	if _, err := s.UpdateAnniversary(context.Background(), 3, 7, draft); !errors.Is(err, domain.ErrInvalidCover) {
		t.Fatalf("reuse err=%v", err)
	}
}
