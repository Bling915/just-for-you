package together

import (
	"context"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

type Service struct {
	repo          Repository
	relationships RelationshipReader
	clock         Clock
}

func NewService(repo Repository, relationships RelationshipReader, clock Clock) *Service {
	return &Service{repo: repo, relationships: relationships, clock: clock}
}
func (s *Service) relationship(ctx context.Context, userID int64) (ActiveRelationship, error) {
	return s.relationships.GetActiveRelationship(ctx, userID)
}
func (s *Service) ListAnniversaries(ctx context.Context, userID int64) ([]domain.Anniversary, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return nil, e
	}
	items, e := s.repo.ListAnniversaries(ctx, r.ID)
	if e != nil {
		return nil, e
	}
	hasBound := false
	for _, item := range items {
		if item.SystemKind() == "relationship" {
			hasBound = true
			break
		}
	}
	if !hasBound && !r.BoundAt.IsZero() {
		items = append(items, domain.RestoreAnniversary(0, r.ID, r.UserAID, "我们在一起", r.BoundAt, true, domain.CountModeElapsed, true, "", "", "relationship", r.BoundAt, r.BoundAt))
	}
	return items, nil
}
func (s *Service) CreateAnniversary(ctx context.Context, userID int64, title string, date time.Time, repeats bool) (domain.Anniversary, error) {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return domain.Anniversary{}, e
	}
	a, e := domain.NewAnniversary(r.ID, userID, title, date, s.clock.Now())
	if e != nil {
		return a, e
	}
	if !repeats {
		a = domain.RestoreAnniversary(a.ID(), a.CoupleID(), a.CreatorID(), a.Title(), a.Date(), false, a.CountMode(), a.Pinned(), a.CoverRef(), a.Message(), a.SystemKind(), a.CreatedAt(), a.UpdatedAt())
	}
	return s.repo.SaveAnniversary(ctx, a)
}
func (s *Service) DeleteAnniversary(ctx context.Context, userID, id int64) error {
	r, e := s.relationship(ctx, userID)
	if e != nil {
		return e
	}
	return s.repo.DeleteAnniversary(ctx, r.ID, id, userID)
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
