package togethermysql

import (
	"context"
	"database/sql"
	"errors"
	app "github.com/nanfang/just-for-you/server/internal/application/together"
	relationship "github.com/nanfang/just-for-you/server/internal/domain/relationship"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) Repository { return Repository{db: db} }
func (r Repository) GetActiveRelationship(ctx context.Context, userID int64) (app.ActiveRelationship, error) {
	var v app.ActiveRelationship
	err := r.db.QueryRowContext(ctx, "SELECT id,user_a_id,user_b_id,bound_at FROM couples WHERE status='active' AND (user_a_id=? OR user_b_id=?)", userID, userID).Scan(&v.ID, &v.UserAID, &v.UserBID, &v.BoundAt)
	if errors.Is(err, sql.ErrNoRows) {
		err = relationship.ErrNoActiveRelationship
	}
	return v, err
}

const anniversaryColumns = "id,couple_id,creator_user_id,title,anniversary_date,repeats_yearly,count_mode,pinned,cover_ref,message,system_kind,created_at,updated_at"

func (r Repository) ListAnniversaries(ctx context.Context, coupleID int64) ([]domain.Anniversary, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT "+anniversaryColumns+" FROM anniversaries WHERE couple_id=? ORDER BY anniversary_date,id", coupleID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Anniversary{}
	for rows.Next() {
		var id, cid, uid int64
		var title, kind, countMode, coverRef, message string
		var date anniversaryDate
		var created, updated time.Time
		var repeats, pinned bool
		if e = rows.Scan(&id, &cid, &uid, &title, &date, &repeats, &countMode, &pinned, &coverRef, &message, &kind, &created, &updated); e != nil {
			return nil, e
		}
		out = append(out, domain.RestoreAnniversary(id, cid, uid, title, date.Time, repeats, domain.CountMode(countMode), pinned, coverRef, message, kind, created, updated))
	}
	return out, rows.Err()
}
func (r Repository) SaveAnniversary(ctx context.Context, v domain.Anniversary) (domain.Anniversary, error) {
	res, e := r.db.ExecContext(ctx, "INSERT INTO anniversaries(couple_id,creator_user_id,title,anniversary_date,repeats_yearly,count_mode,pinned,cover_ref,message,system_kind,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)", v.CoupleID(), v.CreatorID(), v.Title(), toAnniversaryDate(v.Date()), v.RepeatsYearly(), v.CountMode(), v.Pinned(), v.CoverRef(), v.Message(), v.SystemKind(), v.CreatedAt(), v.UpdatedAt())
	if e != nil {
		return v, e
	}
	id, e := res.LastInsertId()
	return domain.RestoreAnniversary(id, v.CoupleID(), v.CreatorID(), v.Title(), v.Date(), v.RepeatsYearly(), v.CountMode(), v.Pinned(), v.CoverRef(), v.Message(), v.SystemKind(), v.CreatedAt(), v.UpdatedAt()), e
}
func (r Repository) FindAnniversary(ctx context.Context, coupleID, id int64) (domain.Anniversary, error) {
	var aid, cid, uid int64
	var title, mode, cover, message, kind string
	var date anniversaryDate
	var created, updated time.Time
	var repeats, pinned bool
	err := r.db.QueryRowContext(ctx, "SELECT "+anniversaryColumns+" FROM anniversaries WHERE couple_id=? AND id=?", coupleID, id).Scan(&aid, &cid, &uid, &title, &date, &repeats, &mode, &pinned, &cover, &message, &kind, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Anniversary{}, err
	}
	return domain.RestoreAnniversary(aid, cid, uid, title, date.Time, repeats, domain.CountMode(mode), pinned, cover, message, kind, created, updated), nil
}
func (r Repository) UpdateAnniversary(ctx context.Context, v domain.Anniversary) (domain.Anniversary, error) {
	res, err := r.db.ExecContext(ctx, "UPDATE anniversaries SET title=?,anniversary_date=?,repeats_yearly=?,count_mode=?,pinned=?,cover_ref=?,message=?,updated_at=? WHERE couple_id=? AND id=?", v.Title(), toAnniversaryDate(v.Date()), v.RepeatsYearly(), v.CountMode(), v.Pinned(), v.CoverRef(), v.Message(), v.UpdatedAt(), v.CoupleID(), v.ID())
	if err != nil {
		return domain.Anniversary{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return domain.Anniversary{}, err
	}
	if n == 0 {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	return v, nil
}
func (r Repository) CreateCoverGrant(ctx context.Context, g app.CoverUploadGrant) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO upload_grants(user_id,object_key,content_type,purpose,resource_id,expires_at) VALUES(?,?,?,?,?,?)", g.OwnerID, g.ObjectKey, g.ContentType, g.Purpose, g.ResourceID, g.ExpiresAt)
	return err
}
func (r Repository) GetCoverGrant(ctx context.Context, key string) (app.CoverUploadGrant, error) {
	var g app.CoverUploadGrant
	var consumed sql.NullTime
	err := r.db.QueryRowContext(ctx, "SELECT object_key,user_id,resource_id,content_type,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=?", key).Scan(&g.ObjectKey, &g.OwnerID, &g.ResourceID, &g.ContentType, &g.Purpose, &g.ExpiresAt, &consumed)
	if errors.Is(err, sql.ErrNoRows) {
		return g, domain.ErrInvalidCover
	}
	if consumed.Valid {
		g.ConsumedAt = &consumed.Time
	}
	return g, err
}
func (r Repository) UpdateAnniversaryWithCoverGrant(ctx context.Context, v domain.Anniversary, key string, now time.Time) (domain.Anniversary, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Anniversary{}, err
	}
	defer tx.Rollback()
	var owner, resource int64
	var purpose string
	var expires time.Time
	var consumed sql.NullTime
	if err = tx.QueryRowContext(ctx, "SELECT user_id,resource_id,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE", key).Scan(&owner, &resource, &purpose, &expires, &consumed); err != nil || owner == 0 || resource != v.ID() || purpose != "anniversary_cover" || consumed.Valid || !expires.After(now) {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	res, err := tx.ExecContext(ctx, "UPDATE anniversaries SET title=?,anniversary_date=?,repeats_yearly=?,count_mode=?,pinned=?,cover_ref=?,message=?,updated_at=? WHERE couple_id=? AND id=?", v.Title(), toAnniversaryDate(v.Date()), v.RepeatsYearly(), v.CountMode(), v.Pinned(), v.CoverRef(), v.Message(), v.UpdatedAt(), v.CoupleID(), v.ID())
	if err != nil {
		return domain.Anniversary{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	if _, err = tx.ExecContext(ctx, "UPDATE upload_grants SET consumed_at=? WHERE object_key=?", now, key); err != nil {
		return domain.Anniversary{}, err
	}
	return v, tx.Commit()
}
func (r Repository) DeleteAnniversary(ctx context.Context, coupleID, id, userID int64) error {
	return deleteOwned(ctx, r.db, "anniversaries", coupleID, id, userID, " AND system_kind='none'")
}
func (r Repository) ListTodos(ctx context.Context, coupleID int64) ([]domain.Todo, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,couple_id,creator_user_id,completed_by_user_id,title,due_at,status,completed_at,created_at,updated_at FROM todos WHERE couple_id=? ORDER BY status,due_at IS NULL,due_at,id DESC", coupleID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Todo{}
	for rows.Next() {
		v, e := scanTodo(rows.Scan)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r Repository) SaveTodo(ctx context.Context, v domain.Todo) (domain.Todo, error) {
	res, e := r.db.ExecContext(ctx, "INSERT INTO todos(couple_id,creator_user_id,title,due_at,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?)", v.CoupleID(), v.CreatorID(), v.Title(), v.DueAt(), v.Status(), v.CreatedAt(), v.UpdatedAt())
	if e != nil {
		return v, e
	}
	id, e := res.LastInsertId()
	return domain.RestoreTodo(id, v.CoupleID(), v.CreatorID(), nil, v.Title(), v.DueAt(), v.Status(), nil, v.CreatedAt(), v.UpdatedAt()), e
}
func (r Repository) FindTodo(ctx context.Context, coupleID, id int64) (domain.Todo, error) {
	v, e := scanTodo(r.db.QueryRowContext(ctx, "SELECT id,couple_id,creator_user_id,completed_by_user_id,title,due_at,status,completed_at,created_at,updated_at FROM todos WHERE couple_id=? AND id=?", coupleID, id).Scan)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return v, e
}
func (r Repository) UpdateTodo(ctx context.Context, v domain.Todo) (domain.Todo, error) {
	_, e := r.db.ExecContext(ctx, "UPDATE todos SET status=?,completed_by_user_id=?,completed_at=?,updated_at=? WHERE couple_id=? AND id=?", v.Status(), v.CompletedBy(), v.CompletedAt(), v.UpdatedAt(), v.CoupleID(), v.ID())
	return v, e
}
func (r Repository) DeleteTodo(ctx context.Context, coupleID, id, userID int64) error {
	return deleteOwned(ctx, r.db, "todos", coupleID, id, userID, "")
}
func (r Repository) ListWishes(ctx context.Context, coupleID int64) ([]domain.Wish, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,couple_id,creator_user_id,completed_by_user_id,linked_moment_id,title,description,status,completed_at,created_at,updated_at FROM wishes WHERE couple_id=? ORDER BY status,id DESC", coupleID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Wish{}
	for rows.Next() {
		v, e := scanWish(rows.Scan)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r Repository) SaveWish(ctx context.Context, v domain.Wish) (domain.Wish, error) {
	res, e := r.db.ExecContext(ctx, "INSERT INTO wishes(couple_id,creator_user_id,title,description,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?)", v.CoupleID(), v.CreatorID(), v.Title(), v.Description(), v.Status(), v.CreatedAt(), v.UpdatedAt())
	if e != nil {
		return v, e
	}
	id, e := res.LastInsertId()
	return domain.RestoreWish(id, v.CoupleID(), v.CreatorID(), nil, nil, v.Title(), v.Description(), v.Status(), nil, v.CreatedAt(), v.UpdatedAt()), e
}
func (r Repository) FindWish(ctx context.Context, coupleID, id int64) (domain.Wish, error) {
	v, e := scanWish(r.db.QueryRowContext(ctx, "SELECT id,couple_id,creator_user_id,completed_by_user_id,linked_moment_id,title,description,status,completed_at,created_at,updated_at FROM wishes WHERE couple_id=? AND id=?", coupleID, id).Scan)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return v, e
}
func (r Repository) UpdateWish(ctx context.Context, v domain.Wish) (domain.Wish, error) {
	_, e := r.db.ExecContext(ctx, "UPDATE wishes SET status=?,completed_by_user_id=?,linked_moment_id=?,completed_at=?,updated_at=? WHERE couple_id=? AND id=?", v.Status(), v.CompletedBy(), v.LinkedMomentID(), v.CompletedAt(), v.UpdatedAt(), v.CoupleID(), v.ID())
	return v, e
}
func (r Repository) DeleteWish(ctx context.Context, coupleID, id, userID int64) error {
	return deleteOwned(ctx, r.db, "wishes", coupleID, id, userID, "")
}
func (r Repository) MomentCoupleID(ctx context.Context, id int64) (int64, error) {
	var cid int64
	e := r.db.QueryRowContext(ctx, "SELECT couple_id FROM moments WHERE id=? AND deleted_at IS NULL", id).Scan(&cid)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrInvalidMoment
	}
	return cid, e
}

type scanner func(...any) error

func scanTodo(scan scanner) (domain.Todo, error) {
	var id, cid, uid int64
	var by sql.NullInt64
	var title, status string
	var due, done sql.NullTime
	var created, updated time.Time
	e := scan(&id, &cid, &uid, &by, &title, &due, &status, &done, &created, &updated)
	if e != nil {
		return domain.Todo{}, e
	}
	return domain.RestoreTodo(id, cid, uid, intPtr(by), title, timePtr(due), status, timePtr(done), created, updated), nil
}
func scanWish(scan scanner) (domain.Wish, error) {
	var id, cid, uid int64
	var by, moment sql.NullInt64
	var title, description, status string
	var done sql.NullTime
	var created, updated time.Time
	e := scan(&id, &cid, &uid, &by, &moment, &title, &description, &status, &done, &created, &updated)
	if e != nil {
		return domain.Wish{}, e
	}
	return domain.RestoreWish(id, cid, uid, intPtr(by), intPtr(moment), title, description, status, timePtr(done), created, updated), nil
}
func intPtr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	x := v.Int64
	return &x
}
func timePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	x := v.Time
	return &x
}
func deleteOwned(ctx context.Context, db *sql.DB, table string, coupleID, id, userID int64, suffix string) error {
	q := "DELETE FROM " + table + " WHERE couple_id=? AND id=? AND creator_user_id=?" + suffix
	res, e := db.ExecContext(ctx, q, coupleID, id, userID)
	if e != nil {
		return e
	}
	n, e := res.RowsAffected()
	if e == nil && n == 0 {
		return domain.ErrNotCreator
	}
	return e
}
