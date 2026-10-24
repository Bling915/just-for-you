package anniversarymysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	app "github.com/nanfang/just-for-you/server/internal/anniversary/application"
	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
)

type MySQLAnniversaryStore struct {
	db *sql.DB
}

func NewMySQLAnniversaryStore(db *sql.DB) MySQLAnniversaryStore {
	return MySQLAnniversaryStore{db: db}
}

const anniversaryColumns = "id,couple_id,creator_user_id,title,anniversary_date,repeats_yearly,count_mode,pinned,cover_ref,message,system_kind,created_at,updated_at"

func (store MySQLAnniversaryStore) List(
	ctx context.Context,
	coupleID int64,
) ([]domain.Anniversary, error) {
	rows, err := store.db.QueryContext(
		ctx,
		"SELECT "+anniversaryColumns+" FROM anniversaries WHERE couple_id=? ORDER BY anniversary_date,id",
		coupleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Anniversary{}
	for rows.Next() {
		item, err := scanAnniversary(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store MySQLAnniversaryStore) Save(
	ctx context.Context,
	item domain.Anniversary,
) (domain.Anniversary, error) {
	result, err := store.db.ExecContext(
		ctx,
		"INSERT INTO anniversaries(couple_id,creator_user_id,title,anniversary_date,repeats_yearly,count_mode,pinned,cover_ref,message,system_kind,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
		item.CoupleID(),
		item.CreatorID(),
		item.Title(),
		toAnniversaryDate(item.Date()),
		item.RepeatsYearly(),
		item.CountMode(),
		item.Pinned(),
		item.CoverRef(),
		item.Message(),
		item.SystemKind(),
		item.CreatedAt(),
		item.UpdatedAt(),
	)
	if err != nil {
		return item, err
	}
	id, err := result.LastInsertId()
	return domain.RestoreAnniversary(
		id,
		item.CoupleID(),
		item.CreatorID(),
		item.Title(),
		item.Date(),
		item.RepeatsYearly(),
		item.CountMode(),
		item.Pinned(),
		item.CoverRef(),
		item.Message(),
		item.SystemKind(),
		item.CreatedAt(),
		item.UpdatedAt(),
	), err
}

func (store MySQLAnniversaryStore) Find(
	ctx context.Context,
	coupleID, id int64,
) (domain.Anniversary, error) {
	item, err := scanAnniversary(
		store.db.QueryRowContext(
			ctx,
			"SELECT "+anniversaryColumns+" FROM anniversaries WHERE couple_id=? AND id=?",
			coupleID,
			id,
		).Scan,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	return item, err
}

func (store MySQLAnniversaryStore) Update(
	ctx context.Context,
	item domain.Anniversary,
) (domain.Anniversary, error) {
	result, err := store.db.ExecContext(
		ctx,
		"UPDATE anniversaries SET title=?,anniversary_date=?,repeats_yearly=?,count_mode=?,pinned=?,cover_ref=?,message=?,updated_at=? WHERE couple_id=? AND id=?",
		item.Title(),
		toAnniversaryDate(item.Date()),
		item.RepeatsYearly(),
		item.CountMode(),
		item.Pinned(),
		item.CoverRef(),
		item.Message(),
		item.UpdatedAt(),
		item.CoupleID(),
		item.ID(),
	)
	if err != nil {
		return domain.Anniversary{}, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return domain.Anniversary{}, err
	}
	if count == 0 {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	return item, nil
}

func (store MySQLAnniversaryStore) CreateCoverGrant(
	ctx context.Context,
	grant app.CoverUploadGrant,
) error {
	_, err := store.db.ExecContext(
		ctx,
		"INSERT INTO upload_grants(user_id,object_key,content_type,purpose,resource_id,expires_at) VALUES(?,?,?,?,?,?)",
		grant.OwnerID,
		grant.ObjectKey,
		grant.ContentType,
		grant.Purpose,
		grant.ResourceID,
		grant.ExpiresAt,
	)
	return err
}

func (store MySQLAnniversaryStore) GetCoverGrant(
	ctx context.Context,
	key string,
) (app.CoverUploadGrant, error) {
	var grant app.CoverUploadGrant
	var consumed sql.NullTime
	err := store.db.QueryRowContext(
		ctx,
		"SELECT object_key,user_id,resource_id,content_type,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=?",
		key,
	).Scan(
		&grant.ObjectKey,
		&grant.OwnerID,
		&grant.ResourceID,
		&grant.ContentType,
		&grant.Purpose,
		&grant.ExpiresAt,
		&consumed,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return grant, domain.ErrInvalidCover
	}
	if consumed.Valid {
		grant.ConsumedAt = &consumed.Time
	}
	return grant, err
}

func (store MySQLAnniversaryStore) UpdateWithCoverGrant(
	ctx context.Context,
	item domain.Anniversary,
	key string,
	now time.Time,
) (domain.Anniversary, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Anniversary{}, err
	}
	defer tx.Rollback()
	var owner, resource int64
	var purpose string
	var expires time.Time
	var consumed sql.NullTime
	err = tx.QueryRowContext(
		ctx,
		"SELECT user_id,resource_id,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE",
		key,
	).Scan(&owner, &resource, &purpose, &expires, &consumed)
	if err != nil ||
		owner == 0 ||
		resource != item.ID() ||
		purpose != "anniversary_cover" ||
		consumed.Valid ||
		!expires.After(now) {
		return domain.Anniversary{}, domain.ErrInvalidCover
	}
	result, err := tx.ExecContext(
		ctx,
		"UPDATE anniversaries SET title=?,anniversary_date=?,repeats_yearly=?,count_mode=?,pinned=?,cover_ref=?,message=?,updated_at=? WHERE couple_id=? AND id=?",
		item.Title(),
		toAnniversaryDate(item.Date()),
		item.RepeatsYearly(),
		item.CountMode(),
		item.Pinned(),
		item.CoverRef(),
		item.Message(),
		item.UpdatedAt(),
		item.CoupleID(),
		item.ID(),
	)
	if err != nil {
		return domain.Anniversary{}, err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return domain.Anniversary{}, domain.ErrNotFound
	}
	if _, err = tx.ExecContext(
		ctx,
		"UPDATE upload_grants SET consumed_at=? WHERE object_key=?",
		now,
		key,
	); err != nil {
		return domain.Anniversary{}, err
	}
	return item, tx.Commit()
}

func (store MySQLAnniversaryStore) Delete(
	ctx context.Context,
	coupleID, id, userID int64,
) error {
	result, err := store.db.ExecContext(
		ctx,
		"DELETE FROM anniversaries WHERE couple_id=? AND id=? AND creator_user_id=? AND system_kind='none'",
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

func scanAnniversary(scan func(...any) error) (domain.Anniversary, error) {
	var id, coupleID, creatorID int64
	var title, countMode, coverRef, message, systemKind string
	var date anniversaryDate
	var repeats, pinned bool
	var created, updated time.Time
	err := scan(
		&id,
		&coupleID,
		&creatorID,
		&title,
		&date,
		&repeats,
		&countMode,
		&pinned,
		&coverRef,
		&message,
		&systemKind,
		&created,
		&updated,
	)
	if err != nil {
		return domain.Anniversary{}, err
	}
	return domain.RestoreAnniversary(
		id,
		coupleID,
		creatorID,
		title,
		date.Time,
		repeats,
		domain.CountMode(countMode),
		pinned,
		coverRef,
		message,
		systemKind,
		created,
		updated,
	), nil
}
