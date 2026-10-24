package dailyquestionmysql

import (
	"context"
	"database/sql"
	"errors"

	app "github.com/nanfang/just-for-you/server/internal/dailyquestion/application"
	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
	momentdomain "github.com/nanfang/just-for-you/server/internal/moment/domain"
)

type MySQLAnswerSubmissionStore struct {
	db        *sql.DB
	questions app.QuestionStore
}

func NewMySQLAnswerSubmissionStore(
	db *sql.DB,
	questions app.QuestionStore,
) MySQLAnswerSubmissionStore {
	return MySQLAnswerSubmissionStore{db: db, questions: questions}
}

func (store MySQLAnswerSubmissionStore) Submit(
	ctx context.Context,
	command app.SubmitAnswerTransaction,
) (domain.DailyQuestion, error) {
	answerText, err := domain.ValidateAnswer(command.Text)
	if err != nil || command.UserID == command.PartnerID {
		return domain.DailyQuestion{}, domain.ErrInvalidDailyAnswer
	}
	tx, err := store.db.BeginTx(ctx, nil)
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
		command.QuestionID,
		command.UserID,
		command.UserID,
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
		command.QuestionID,
		command.UserID,
		answerText,
		command.Now,
		command.Now,
	)
	if err != nil {
		return domain.DailyQuestion{}, err
	}
	var partnerCount int
	if err = tx.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM daily_answers WHERE daily_question_id=? AND user_id=?",
		command.QuestionID,
		command.PartnerID,
	).Scan(&partnerCount); err != nil {
		return domain.DailyQuestion{}, err
	}
	if partnerCount > 0 {
		if _, err = tx.ExecContext(
			ctx,
			"UPDATE daily_questions SET unlocked_at=?,updated_at=? WHERE id=? AND unlocked_at IS NULL",
			command.Now,
			command.Now,
			command.QuestionID,
		); err != nil {
			return domain.DailyQuestion{}, err
		}
		projection, err := projectionText(ctx, tx, command.QuestionID, questionText)
		if err != nil {
			return domain.DailyQuestion{}, err
		}
		if _, err = momentdomain.NewStatus(""); err != nil {
			return domain.DailyQuestion{}, domain.ErrInvalidDailyAnswer
		}
		if _, err = tx.ExecContext(
			ctx,
			"INSERT IGNORE INTO moments(couple_id,author_user_id,space_type,visibility,daily_question_id,body,status,place_name,occurred_at,created_at,updated_at) VALUES(?,?,'shared','couple',?,?,'','',?,?,?)",
			coupleID,
			command.UserID,
			command.QuestionID,
			projection,
			command.Now,
			command.Now,
			command.Now,
		); err != nil {
			return domain.DailyQuestion{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return domain.DailyQuestion{}, err
	}
	return store.questions.FindByID(ctx, coupleID, command.QuestionID)
}

func projectionText(
	ctx context.Context,
	tx *sql.Tx,
	questionID int64,
	questionText string,
) (string, error) {
	rows, err := tx.QueryContext(
		ctx,
		"SELECT answer_text FROM daily_answers WHERE daily_question_id=? ORDER BY user_id",
		questionID,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	body := "每日一句 · " + questionText
	for rows.Next() {
		var answer string
		if err = rows.Scan(&answer); err != nil {
			return "", err
		}
		body += "\n" + answer
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	projection, err := momentdomain.ValidateText(body)
	if err != nil {
		return "", domain.ErrInvalidDailyAnswer
	}
	return projection, nil
}
