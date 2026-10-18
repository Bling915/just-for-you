package infrastructure

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAcceptInvitationUsesLockAndTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT code,creator_user_id,expires_at,consumed_at FROM invite_codes WHERE code=? FOR UPDATE")).WithArgs("AB3K9Z").WillReturnRows(sqlmock.NewRows([]string{"code", "creator_user_id", "expires_at", "consumed_at"}).AddRow("AB3K9Z", int64(1), now.Add(time.Minute), nil))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM couples").WithArgs(int64(2), int64(1), int64(2), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("INSERT INTO couples").WithArgs(int64(1), int64(2), now).WillReturnResult(sqlmock.NewResult(9, 1))
	mock.ExpectExec("UPDATE invite_codes SET consumed_at").WithArgs(now, int64(2), "AB3K9Z").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	relationship, err := NewMySQLRepository(db).AcceptInvitation(context.Background(), "AB3K9Z", 2, now)
	if err != nil {
		t.Fatal(err)
	}
	if relationship.ID() != 9 {
		t.Fatalf("id=%d", relationship.ID())
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
