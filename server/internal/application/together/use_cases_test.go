package together

import (
	"context"
	"errors"
	relationship "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	"testing"
	"time"
)

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

type fakeRelations struct {
	rel ActiveRelationship
	err error
}

func (f fakeRelations) GetActiveRelationship(context.Context, int64) (ActiveRelationship, error) {
	return f.rel, f.err
}

type fakeRepo struct {
	todo    domain.Todo
	updated int
}

func (f *fakeRepo) ListAnniversaries(context.Context, int64) ([]domain.Anniversary, error) {
	return nil, nil
}
func (f *fakeRepo) SaveAnniversary(_ context.Context, v domain.Anniversary) (domain.Anniversary, error) {
	return v, nil
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
func TestEndedRelationshipDeniesTogetherData(t *testing.T) {
	s := NewService(&fakeRepo{}, fakeRelations{err: relationship.ErrNoActiveRelationship}, fakeClock{})
	if _, e := s.ListTodos(context.Background(), 2); !errors.Is(e, relationship.ErrNoActiveRelationship) {
		t.Fatalf("got %v", e)
	}
}
func TestTodoCompletionIsIdempotent(t *testing.T) {
	now := time.Now()
	todo, _ := domain.NewTodo(1, 2, "散步", nil, now)
	repo := &fakeRepo{todo: todo}
	s := NewService(repo, fakeRelations{rel: ActiveRelationship{ID: 1, UserAID: 2, UserBID: 3}}, fakeClock{now: now})
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
