package togethermysql

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	app "github.com/nanfang/just-for-you/server/internal/application/together"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return Repository{db: db}
}

func (repository Repository) AIPersonalizationEnabled(
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

func (repository Repository) RecentContextMoments(
	ctx context.Context,
	coupleID int64,
	limit int,
) ([]app.ContextMoment, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		"SELECT body,status_code,place_name,space_type,visibility,occurred_at FROM moments WHERE couple_id=? AND space_type='shared' AND visibility='couple' AND deleted_at IS NULL ORDER BY occurred_at DESC LIMIT ?",
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

func (repository Repository) Find(
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

func (repository Repository) FindByID(
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

func (repository Repository) scanDailyQuestion(
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

func (repository Repository) Create(
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

func (repository Repository) SubmitAnswer(
	ctx context.Context,
	questionID, userID, partnerID int64,
	text string,
	now time.Time,
) (domain.DailyQuestion, error) {
	answerText := strings.TrimSpace(text)
	if answerText == "" || len([]rune(answerText)) > 500 {
		return domain.DailyQuestion{}, domain.ErrInvalidDailyAnswer
	}
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	defer tx.Rollback()
	var coupleID int64
	var questionText string
	var unlocked sql.NullTime
	err = tx.QueryRowContext(
		ctx,
		"SELECT q.couple_id,q.question_text,q.unlocked_at FROM daily_questions q JOIN couples c ON c.id=q.couple_id WHERE q.id=? AND c.status='active' AND (c.user_a_id=? OR c.user_b_id=?) FOR UPDATE",
		questionID,
		userID,
		userID,
	).Scan(&coupleID, &questionText, &unlocked)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DailyQuestion{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	if unlocked.Valid {
		return domain.DailyQuestion{}, domain.ErrDailyQuestionFrozen
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO daily_answers(daily_question_id,user_id,answer_text,submitted_at,updated_at) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE answer_text=VALUES(answer_text),submitted_at=VALUES(submitted_at),updated_at=VALUES(updated_at)",
		questionID,
		userID,
		answerText,
		now,
		now,
	)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	var partnerCount int
	if err = tx.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM daily_answers WHERE daily_question_id=? AND user_id=?",
		questionID,
		partnerID,
	).Scan(&partnerCount); err != nil {
		return domain.DailyQuestion{}, err
	}
	if partnerCount > 0 {
		if _, err = tx.ExecContext(
			ctx,
			"UPDATE daily_questions SET unlocked_at=?,updated_at=? WHERE id=? AND unlocked_at IS NULL",
			now,
			now,
			questionID,
		); err != nil {
			return domain.DailyQuestion{}, err
		}
		rows, err := tx.QueryContext(
			ctx,
			"SELECT answer_text FROM daily_answers WHERE daily_question_id=? ORDER BY user_id",
			questionID,
		)
		if err != nil {
			return domain.DailyQuestion{}, err
		}
		body := "每日一句 · " + questionText
		for rows.Next() {
			var answer string
			if err = rows.Scan(&answer); err != nil {
				rows.Close()
				return domain.DailyQuestion{}, err
			}
			body += "\n" + answer
		}
		rows.Close()
		if _, err = tx.ExecContext(
			ctx,
			"INSERT IGNORE INTO moments(couple_id,author_user_id,space_type,visibility,kind,daily_question_id,body,status_code,place_name,occurred_at,created_at,updated_at) VALUES(?,?,'shared','couple','daily_question',?,?,'','',?,?,?)",
			coupleID,
			userID,
			questionID,
			body,
			now,
			now,
			now,
		); err != nil {
			return domain.DailyQuestion{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.DailyQuestion{}, err
	}
	return repository.FindByID(ctx, coupleID, questionID)
}

func (repository Repository) List(
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
