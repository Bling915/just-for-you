package wishmysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/wish/domain"
)

type MySQLWishStore struct {
	db *sql.DB
}

func NewMySQLWishStore(db *sql.DB) MySQLWishStore {
	return MySQLWishStore{db: db}
}

func (store MySQLWishStore) List(
	ctx context.Context,
	coupleID int64,
) ([]domain.Wish, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT id,couple_id,creator_user_id,completed_by_user_id,linked_moment_id,title,description,status,completed_at,created_at,updated_at FROM wishes WHERE couple_id=? ORDER BY status,id DESC",
		coupleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Wish{}
	for rows.Next() {
		item, err := scanWish(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store MySQLWishStore) Save(
	ctx context.Context,
	item domain.Wish,
) (domain.Wish, error) {
	result, err := store.db.ExecContext(
		ctx,
		"INSERT INTO wishes(couple_id,creator_user_id,title,description,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?)",
		item.CoupleID(),
		item.CreatorID(),
		item.Title(),
		item.Description(),
		item.Status(),
		item.CreatedAt(),
		item.UpdatedAt(),
	)
	if err != nil {
		return item, err
	}
	id, err := result.LastInsertId()
	return domain.RestoreWish(
		id,
		item.CoupleID(),
		item.CreatorID(),
		nil,
		nil,
		item.Title(),
		item.Description(),
		item.Status(),
		nil,
		item.CreatedAt(),
		item.UpdatedAt(),
	), err
}

func (store MySQLWishStore) Find(
	ctx context.Context,
	coupleID, id int64,
) (domain.Wish, error) {
	item, err := scanWish(
		store.db.QueryRowContext(
			ctx,
			"SELECT id,couple_id,creator_user_id,completed_by_user_id,linked_moment_id,title,description,status,completed_at,created_at,updated_at FROM wishes WHERE couple_id=? AND id=?",
			coupleID,
			id,
		).Scan,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Wish{}, domain.ErrNotFound
	}
	return item, err
}

func (store MySQLWishStore) Update(
	ctx context.Context,
	item domain.Wish,
) (domain.Wish, error) {
	_, err := store.db.ExecContext(
		ctx,
		"UPDATE wishes SET status=?,completed_by_user_id=?,linked_moment_id=?,completed_at=?,updated_at=? WHERE couple_id=? AND id=?",
		item.Status(),
		item.CompletedBy(),
		item.LinkedMomentID(),
		item.CompletedAt(),
		item.UpdatedAt(),
		item.CoupleID(),
		item.ID(),
	)
	return item, err
}

func (store MySQLWishStore) Delete(
	ctx context.Context,
	coupleID, id, userID int64,
) error {
	result, err := store.db.ExecContext(
		ctx,
		"DELETE FROM wishes WHERE couple_id=? AND id=? AND creator_user_id=?",
		coupleID,
		id,
		userID,
	)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return domain.ErrNotCreator
	}
	return err
}

type momentCoupleReader struct {
	db *sql.DB
}

func NewMomentCoupleReader(db *sql.DB) momentCoupleReader {
	return momentCoupleReader{db: db}
}

func (reader momentCoupleReader) MomentCoupleID(ctx context.Context, id int64) (int64, error) {
	var coupleID int64
	err := reader.db.QueryRowContext(
		ctx,
		"SELECT couple_id FROM moments WHERE id=? AND deleted_at IS NULL",
		id,
	).Scan(&coupleID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrInvalidMoment
	}
	return coupleID, err
}

func scanWish(scan func(...any) error) (domain.Wish, error) {
	var id, coupleID, creatorID int64
	var completedBy, linkedMomentID sql.NullInt64
	var title, description, status string
	var completedAt sql.NullTime
	var created, updated time.Time
	err := scan(
		&id,
		&coupleID,
		&creatorID,
		&completedBy,
		&linkedMomentID,
		&title,
		&description,
		&status,
		&completedAt,
		&created,
		&updated,
	)
	if err != nil {
		return domain.Wish{}, err
	}
	return domain.RestoreWish(
		id,
		coupleID,
		creatorID,
		nullIntPointer(completedBy),
		nullIntPointer(linkedMomentID),
		title,
		description,
		status,
		nullTimePointer(completedAt),
		created,
		updated,
	), nil
}

func nullIntPointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
