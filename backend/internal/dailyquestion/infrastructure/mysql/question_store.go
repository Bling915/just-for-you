package dailyquestionmysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	app "github.com/nanfang/just-for-you/server/internal/dailyquestion/application"
	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
)

type MySQLQuestionStore struct {
	db *sql.DB
}

func NewMySQLQuestionStore(db *sql.DB) MySQLQuestionStore {
	return MySQLQuestionStore{db: db}
}

func (repository MySQLQuestionStore) AIPersonalizationEnabled(
	ctx context.Context,
	coupleID int64,
) (bool, error) {
	var disabled int
	err := repository.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM notification_preferences WHERE couple_id=? AND ai_personalization_enabled=FALSE",
		coupleID,
	).Scan(&disabled)
	return disabled == 0, err
}

func (repository MySQLQuestionStore) RecentContextMoments(
	ctx context.Context,
	coupleID int64,
	limit int,
) ([]app.ContextMoment, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		"SELECT body,status,place_name,space_type,visibility,occurred_at FROM moments WHERE couple_id=? AND space_type='shared' AND visibility='couple' AND deleted_at IS NULL ORDER BY occurred_at DESC LIMIT ?",
		coupleID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []app.ContextMoment{}
	for rows.Next() {
		var item app.ContextMoment
		if err = rows.Scan(
			&item.Text,
			&item.Status,
			&item.Place,
			&item.SpaceType,
			&item.Visibility,
			&item.OccurredAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

const dailyQuestionColumns = "id,couple_id,question_date,question_text,source,context_version,unlocked_at,created_at,updated_at"

func (repository MySQLQuestionStore) Find(
	ctx context.Context,
	coupleID int64,
	date time.Time,
) (domain.DailyQuestion, error) {
	return repository.scanDailyQuestion(
		ctx,
		repository.db.QueryRowContext(
			ctx,
			"SELECT "+dailyQuestionColumns+" FROM daily_questions WHERE couple_id=? AND question_date=?",
			coupleID,
			toQuestionDate(date),
		).Scan,
	)
}

func (repository MySQLQuestionStore) FindByID(
	ctx context.Context,
	coupleID, id int64,
) (domain.DailyQuestion, error) {
	return repository.scanDailyQuestion(
		ctx,
		repository.db.QueryRowContext(
			ctx,
			"SELECT "+dailyQuestionColumns+" FROM daily_questions WHERE couple_id=? AND id=?",
			coupleID,
			id,
		).Scan,
	)
}

func (repository MySQLQuestionStore) scanDailyQuestion(
	ctx context.Context,
	scan func(...any) error,
) (domain.DailyQuestion, error) {
	var id, coupleID int64
	var date questionDate
	var text, source, version string
	var unlocked sql.NullTime
	var created, updated time.Time
	if err := scan(
		&id,
		&coupleID,
		&date,
		&text,
		&source,
		&version,
		&unlocked,
		&created,
		&updated,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrNotFound
		}
		return domain.DailyQuestion{}, err
	}
	rows, err := repository.db.QueryContext(
		ctx,
		"SELECT user_id,answer_text,submitted_at,updated_at FROM daily_answers WHERE daily_question_id=? ORDER BY user_id",
		id,
	)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	defer rows.Close()
	answers := []domain.DailyAnswer{}
	for rows.Next() {
		var answer domain.DailyAnswer
		if err = rows.Scan(
			&answer.UserID,
			&answer.Text,
			&answer.SubmittedAt,
			&answer.UpdatedAt,
		); err != nil {
			return domain.DailyQuestion{}, err
		}
		answers = append(answers, answer)
	}
	return domain.RestoreDailyQuestion(
		id,
		coupleID,
		date.Time,
		text,
		source,
		version,
		answers,
		timePtr(unlocked),
		created,
		updated,
	), rows.Err()
}

func (repository MySQLQuestionStore) Create(
	ctx context.Context,
	value domain.DailyQuestion,
) (domain.DailyQuestion, error) {
	_, err := repository.db.ExecContext(
		ctx,
		"INSERT IGNORE INTO daily_questions(couple_id,question_date,question_text,source,context_version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)",
		value.CoupleID(),
		toQuestionDate(value.Date()),
		value.Text(),
		value.Source(),
		value.ContextVersion(),
		value.CreatedAt(),
		value.UpdatedAt(),
	)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	return repository.Find(ctx, value.CoupleID(), value.Date())
}

func (repository MySQLQuestionStore) List(
	ctx context.Context,
	coupleID int64,
	limit int,
) ([]domain.DailyQuestion, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		"SELECT id FROM daily_questions WHERE couple_id=? ORDER BY question_date DESC,id DESC LIMIT ?",
		coupleID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	items := make([]domain.DailyQuestion, 0, len(ids))
	for _, id := range ids {
		item, err := repository.FindByID(ctx, coupleID, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
