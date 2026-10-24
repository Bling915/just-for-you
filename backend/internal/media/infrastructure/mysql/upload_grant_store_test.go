package mysql

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nanfang/just-for-you/server/internal/media/domain"
	"regexp"
	"testing"
	"time"
)

func TestUploadGrantStoreConsumesOnlyMatchingUnusedGrant(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id,resource_id,content_type,purpose,expires_at,consumed_at FROM upload_grants WHERE object_key=? FOR UPDATE")).WithArgs("x").WillReturnRows(sqlmock.NewRows([]string{"user_id", "resource_id", "content_type", "purpose", "expires_at", "consumed_at"}).AddRow(7, 9, "image/jpeg", "anniversary_cover", now.Add(time.Minute), nil))
	mock.ExpectExec("UPDATE upload_grants SET consumed_at").WithArgs(now, "x").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err := NewUploadGrantStore(db).Consume(context.Background(), "x", 7, 9, domain.PurposeAnniversaryCover, now)
	if err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
