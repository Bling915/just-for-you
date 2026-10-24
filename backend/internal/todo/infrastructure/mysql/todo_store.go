package todomysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	domain "github.com/nanfang/just-for-you/server/internal/todo/domain"
)

type MySQLTodoStore struct {
	db *sql.DB
}

func NewMySQLTodoStore(db *sql.DB) MySQLTodoStore {
	return MySQLTodoStore{db: db}
}

func (store MySQLTodoStore) List(
	ctx context.Context,
	coupleID int64,
) ([]domain.Todo, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT id,couple_id,creator_user_id,completed_by_user_id,title,due_at,status,completed_at,created_at,updated_at FROM todos WHERE couple_id=? ORDER BY status,due_at IS NULL,due_at,id DESC",
		coupleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Todo{}
	for rows.Next() {
		item, err := scanTodo(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store MySQLTodoStore) Save(
	ctx context.Context,
	item domain.Todo,
) (domain.Todo, error) {
	result, err := store.db.ExecContext(
		ctx,
		"INSERT INTO todos(couple_id,creator_user_id,title,due_at,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?)",
		item.CoupleID(),
		item.CreatorID(),
		item.Title(),
		item.DueAt(),
		item.Status(),
		item.CreatedAt(),
		item.UpdatedAt(),
	)
	if err != nil {
		return item, err
	}
	id, err := result.LastInsertId()
	return domain.RestoreTodo(
		id,
		item.CoupleID(),
		item.CreatorID(),
		nil,
		item.Title(),
		item.DueAt(),
		item.Status(),
		nil,
		item.CreatedAt(),
		item.UpdatedAt(),
	), err
}

func (store MySQLTodoStore) Find(
	ctx context.Context,
	coupleID, id int64,
) (domain.Todo, error) {
	item, err := scanTodo(
		store.db.QueryRowContext(
			ctx,
			"SELECT id,couple_id,creator_user_id,completed_by_user_id,title,due_at,status,completed_at,created_at,updated_at FROM todos WHERE couple_id=? AND id=?",
			coupleID,
			id,
		).Scan,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Todo{}, domain.ErrNotFound
	}
	return item, err
}

func (store MySQLTodoStore) Update(
	ctx context.Context,
	item domain.Todo,
) (domain.Todo, error) {
	_, err := store.db.ExecContext(
		ctx,
		"UPDATE todos SET status=?,completed_by_user_id=?,completed_at=?,updated_at=? WHERE couple_id=? AND id=?",
		item.Status(),
		item.CompletedBy(),
		item.CompletedAt(),
		item.UpdatedAt(),
		item.CoupleID(),
		item.ID(),
	)
	return item, err
}

func (store MySQLTodoStore) Delete(
	ctx context.Context,
	coupleID, id, userID int64,
) error {
	result, err := store.db.ExecContext(
		ctx,
		"DELETE FROM todos WHERE couple_id=? AND id=? AND creator_user_id=?",
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

func scanTodo(scan func(...any) error) (domain.Todo, error) {
	var id, coupleID, creatorID int64
	var completedBy sql.NullInt64
	var title, status string
	var dueAt, completedAt sql.NullTime
	var created, updated time.Time
	err := scan(
		&id,
		&coupleID,
		&creatorID,
		&completedBy,
		&title,
		&dueAt,
		&status,
		&completedAt,
		&created,
		&updated,
	)
	if err != nil {
		return domain.Todo{}, err
	}
	return domain.RestoreTodo(
		id,
		coupleID,
		creatorID,
		intPointer(completedBy),
		title,
		timePointer(dueAt),
		status,
		timePointer(completedAt),
		created,
		updated,
	), nil
}

func intPointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func timePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
