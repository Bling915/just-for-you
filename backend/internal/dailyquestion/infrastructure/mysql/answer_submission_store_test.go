package dailyquestionmysql

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	app "github.com/nanfang/just-for-you/server/internal/dailyquestion/application"
	domain "github.com/nanfang/just-for-you/server/internal/dailyquestion/domain"
)

type questionStoreStub struct{}

func (questionStoreStub) Find(context.Context, int64, time.Time) (domain.DailyQuestion, error) {
	return domain.DailyQuestion{}, nil
}

func (questionStoreStub) FindByID(context.Context, int64, int64) (domain.DailyQuestion, error) {
	return domain.DailyQuestion{}, nil
}

func (questionStoreStub) Create(
	context.Context,
	domain.DailyQuestion,
) (domain.DailyQuestion, error) {
	return domain.DailyQuestion{}, nil
}

func (questionStoreStub) List(context.Context, int64, int) ([]domain.DailyQuestion, error) {
	return nil, nil
}

func TestAnswerSubmissionWritesAnswerUnlockAndSingleProjectionAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	command := app.SubmitAnswerTransaction{
		QuestionID: 3,
		UserID:     7,
		PartnerID:  8,
		Text:       "今天很开心",
		Now:        now,
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT q.couple_id,q.question_text,q.unlocked_at").
		WithArgs(int64(3), int64(7), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"couple_id", "question_text", "unlocked_at"}).AddRow(9, "今天最开心的事？", nil))
	mock.ExpectExec("INSERT INTO daily_answers").
		WithArgs(int64(3), int64(7), "今天很开心", now, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM daily_answers").
		WithArgs(int64(3), int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec("UPDATE daily_questions SET unlocked_at").
		WithArgs(now, now, int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT answer_text FROM daily_answers").
		WillReturnRows(sqlmock.NewRows([]string{"answer_text"}).AddRow("我的回答").AddRow("TA的回答"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO moments")).
		WithArgs(int64(9), int64(7), int64(3), "每日一句 · 今天最开心的事？\n我的回答\nTA的回答", now, now, now).
		WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectCommit()

	store := NewMySQLAnswerSubmissionStore(db, questionStoreStub{})
	if _, err = store.Submit(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAnswerSubmissionRollsBackWhenProjectionFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	command := app.SubmitAnswerTransaction{
		QuestionID: 3,
		UserID:     7,
		PartnerID:  8,
		Text:       "今天很开心",
		Now:        now,
	}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT q.couple_id,q.question_text,q.unlocked_at").
		WillReturnRows(sqlmock.NewRows([]string{"couple_id", "question_text", "unlocked_at"}).AddRow(9, "问题", nil))
	mock.ExpectExec("INSERT INTO daily_answers").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM daily_answers").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectExec("UPDATE daily_questions SET unlocked_at").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT answer_text FROM daily_answers").
		WillReturnRows(sqlmock.NewRows([]string{"answer_text"}).AddRow("我的回答").AddRow("TA的回答"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT IGNORE INTO moments")).WillReturnError(errors.New("projection failed"))
	mock.ExpectRollback()

	store := NewMySQLAnswerSubmissionStore(db, questionStoreStub{})
	if _, err = store.Submit(context.Background(), command); err == nil {
		t.Fatal("expected projection failure")
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
