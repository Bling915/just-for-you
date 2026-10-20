package infrastructure

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	application "github.com/nanfang/just-for-you/server/internal/application/moment"
	domain "github.com/nanfang/just-for-you/server/internal/domain/moment"
	"regexp"
	"testing"
	"time"
)

func TestSaveWithGrantsPersistsPhotoOrderAndConsumesGrantAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	moment, _ := domain.NewMoment(1, 9, "hello", domain.StatusNone, "", []string{"a", "b"}, now)
	grants := []application.UploadGrant{{ObjectKey: "a", OwnerID: 1, ExpiresAt: now.Add(time.Minute)}, {ObjectKey: "b", OwnerID: 1, ExpiresAt: now.Add(time.Minute)}}
	mock.ExpectBegin()
	for _, key := range []string{"a", "b"} {
		mock.ExpectQuery("SELECT user_id,expires_at,consumed_at FROM upload_grants").WithArgs(key).WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "consumed_at"}).AddRow(1, now.Add(time.Minute), nil))
	}
	mock.ExpectExec("INSERT INTO moments").WithArgs(int64(9), int64(1), "hello", string(domain.StatusNone), "", nil, nil, now, now, now).WillReturnResult(sqlmock.NewResult(20, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO moment_photos(moment_id,object_key,sort_order) VALUES(?,?,?)")).WithArgs(int64(20), "a", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO moment_photos(moment_id,object_key,sort_order) VALUES(?,?,?)")).WithArgs(int64(20), "b", 2).WillReturnResult(sqlmock.NewResult(2, 1))
	for _, key := range []string{"a", "b"} {
		mock.ExpectExec("UPDATE upload_grants SET consumed_at").WithArgs(now, key).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()
	saved, err := NewMySQLRepository(db).SaveWithGrants(context.Background(), moment, grants)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID() != 20 {
		t.Fatalf("id=%d", saved.ID())
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestUpdateWithGrantsUsesOptimisticLockAndUpdatesPhotosAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	version := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	now := version.Add(time.Hour)
	moment := domain.RestoreMomentDetails(20, 1, 9, "edited", domain.StatusMissYou, "外滩", nil, nil, []domain.Photo{domain.NewPhoto("new", 1)}, now.Add(-time.Minute), version.Add(-time.Hour), now, nil)
	grants := []application.UploadGrant{{ObjectKey: "new", OwnerID: 1, ExpiresAt: now.Add(time.Minute)}}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id,expires_at,consumed_at FROM upload_grants").WithArgs("new").WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "consumed_at"}).AddRow(1, now.Add(time.Minute), nil))
	mock.ExpectExec("UPDATE moments SET").WithArgs("edited", string(domain.StatusMissYou), "外滩", nil, nil, moment.OccurredAt(), moment.UpdatedAt(), int64(20), int64(9), version).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM moment_photos").WithArgs(int64(20)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO moment_photos").WithArgs(int64(20), "new", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE upload_grants SET consumed_at").WithArgs(moment.UpdatedAt(), "new").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	updated, err := NewMySQLRepository(db).UpdateWithGrants(context.Background(), moment, version, grants)
	if err != nil || updated.Text() != "edited" {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateWithGrantsReturnsConflictWithoutReplacingPhotos(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	moment, _ := domain.NewMoment(1, 9, "edited", domain.StatusNone, "", nil, now)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE moments SET").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	_, err = NewMySQLRepository(db).UpdateWithGrants(context.Background(), moment, now.Add(-time.Minute), nil)
	if !errors.Is(err, domain.ErrMomentConflict) {
		t.Fatalf("err=%v", err)
	}
}
func TestCursorRoundTripPreservesTimestampAndID(t *testing.T) {
	at := time.Date(2026, 10, 18, 10, 11, 12, 123000000, time.UTC)
	raw := encodeCursor(at, 42)
	gotAt, gotID, err := decodeCursor(raw)
	if err != nil || !gotAt.Equal(at) || gotID != 42 {
		t.Fatalf("at=%s id=%d err=%v", gotAt, gotID, err)
	}
}
