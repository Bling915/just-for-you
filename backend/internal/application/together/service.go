package together

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

type Service struct {
	repo          Repository
	relationships CoupleReader
	clock         Clock
	storage       MediaStorage
	key           func() string
}

func NewServiceWithStorage(repo Repository, relationships CoupleReader, clock Clock, storage MediaStorage, key func() string) *Service {
	return &Service{repo: repo, relationships: relationships, clock: clock, storage: storage, key: key}
}

func NewService(repo Repository, relationships CoupleReader, clock Clock) *Service {
	return &Service{repo: repo, relationships: relationships, clock: clock}
}
func (s *Service) relationship(ctx context.Context, userID int64) (ActiveCouple, error) {
	return s.relationships.GetActiveCouple(ctx, userID)
}
func (s *Service) ListTodos(ctx context.Context, userID int64) ([]domain.Todo, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return nil, e
	}
	return s.repo.ListTodos(ctx, r.ID)
}
func (s *Service) CreateTodo(ctx context.Context, userID int64, title string, due *time.Time) (domain.Todo, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return domain.Todo{}, e
	}
	v, e := domain.NewTodo(r.ID, userID, title, due, s.clock.Now())
	if e != nil {
		return v, e
	}
	return s.repo.SaveTodo(ctx, v)
}
func (s *Service) SetTodoCompleted(ctx context.Context, userID, id int64, completed bool) (domain.Todo, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return domain.Todo{}, e
	}
	v, e := s.repo.FindTodo(ctx, r.ID, id)
	if e != nil {
		return v, e
	}
	if completed {
		v = v.Complete(userID, s.clock.Now())
	} else {
		v = v.Reopen(s.clock.Now())
	}
	return s.repo.UpdateTodo(ctx, v)
}
func (s *Service) DeleteTodo(ctx context.Context, userID, id int64) error {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return e
	}
	return s.repo.DeleteTodo(ctx, r.ID, id, userID)
}
func (s *Service) ListWishes(ctx context.Context, userID int64) ([]domain.Wish, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return nil, e
	}
	return s.repo.ListWishes(ctx, r.ID)
}
func (s *Service) CreateWish(ctx context.Context, userID int64, title, description string) (domain.Wish, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return domain.Wish{}, e
	}
	v, e := domain.NewWish(r.ID, userID, title, description, s.clock.Now())
	if e != nil {
		return v, e
	}
	return s.repo.SaveWish(ctx, v)
}
func (s *Service) SetWishCompleted(ctx context.Context, userID, id int64, completed bool, momentID *int64) (domain.Wish, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return domain.Wish{}, e
	}
	v, e := s.repo.FindWish(ctx, r.ID, id)
	if e != nil {
		return v, e
	}
	if !completed {
		return s.repo.UpdateWish(ctx, v.Reopen(s.clock.Now()))
	}
	momentCouple := r.ID
	if momentID != nil {
		momentCouple, e = s.repo.MomentCoupleID(ctx, *momentID)
		if e != nil {
			return v, e
		}
	}
	v, e = v.Complete(userID, momentID, momentCouple, s.clock.Now())
	if e != nil {
		return v, e
	}
	return s.repo.UpdateWish(ctx, v)
}
func (s *Service) DeleteWish(ctx context.Context, userID, id int64) error {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return e
	}
	return s.repo.DeleteWish(ctx, r.ID, id, userID)
}
