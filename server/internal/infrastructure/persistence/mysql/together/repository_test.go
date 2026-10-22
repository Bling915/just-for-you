package togethermysql

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/together"
)

func TestFindAnniversaryMapsAllFieldsAndScopesCouple(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT id,couple_id,creator_user_id,title,anniversary_date,repeats_yearly,count_mode,pinned,cover_ref,message,system_kind,created_at,updated_at FROM anniversaries").
		WithArgs(int64(9), int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "couple_id", "creator_user_id", "title", "anniversary_date", "repeats_yearly", "count_mode", "pinned", "cover_ref", "message", "system_kind", "created_at", "updated_at"}).
			AddRow(7, 9, 2, "相识", now, true, "elapsed", true, "cover", "永远", "none", now, now))
	got, err := NewRepository(db).FindAnniversary(context.Background(), 9, 7)
	if err != nil || got.CountMode() != domain.CountModeElapsed || !got.Pinned() || got.CoverRef() != "cover" || got.Message() != "永远" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindAnniversaryMapsMissingOrForeignToNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT .* FROM anniversaries").WithArgs(int64(9), int64(70)).WillReturnError(sqlmock.ErrCancelled)
	_, err := NewRepository(db).FindAnniversary(context.Background(), 9, 70)
	if errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-row database error must not become not found: %v", err)
	}

	db2, mock2, _ := sqlmock.New()
	defer db2.Close()
	mock2.ExpectQuery("SELECT .* FROM anniversaries").WithArgs(int64(9), int64(71)).WillReturnRows(sqlmock.NewRows([]string{"id", "couple_id", "creator_user_id", "title", "anniversary_date", "repeats_yearly", "count_mode", "pinned", "cover_ref", "message", "system_kind", "created_at", "updated_at"}))
	_, err = NewRepository(db2).FindAnniversary(context.Background(), 9, 71)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error=%v", err)
	}
}

func TestUpdateAnniversaryPersistsEveryEditableField(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	a := domain.RestoreAnniversary(7, 9, 2, "旅行", now, false, domain.CountModeElapsed, true, "cover", "出发", "none", now, now)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE anniversaries SET title=?,anniversary_date=?,repeats_yearly=?,count_mode=?,pinned=?,cover_ref=?,message=?,updated_at=? WHERE couple_id=? AND id=?")).
		WithArgs("旅行", now, false, "elapsed", true, "cover", "出发", now, int64(9), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := NewRepository(db).UpdateAnniversary(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
