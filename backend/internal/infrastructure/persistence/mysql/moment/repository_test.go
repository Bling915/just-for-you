package infrastructure

import (
	"context"
	"database/sql/driver"
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
	mock.ExpectQuery("SELECT id FROM couples").WithArgs(int64(9), int64(1), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectExec("INSERT INTO moments").WithArgs(int64(9), int64(1), string(domain.SpaceShared), string(domain.VisibilityCouple), string(domain.KindStandard), nil, nil, "hello", string(domain.StatusNone), "", nil, nil, now, now, now).WillReturnResult(sqlmock.NewResult(20, 1))
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
	moment := domain.RestoreMomentDetails(domain.RestoreInput{ID: 20, AuthorID: 1, RelationshipID: 9, Text: "edited", Status: domain.StatusMissYou, Place: "外滩", Photos: []domain.Photo{domain.NewPhoto("new", 1)}, OccurredAt: now.Add(-time.Minute), CreatedAt: version.Add(-time.Hour), UpdatedAt: now, SpaceType: domain.SpaceShared, Visibility: domain.VisibilityCouple, Kind: domain.KindStandard})
	grants := []application.UploadGrant{{ObjectKey: "new", OwnerID: 1, ExpiresAt: now.Add(time.Minute)}}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id,expires_at,consumed_at FROM upload_grants").WithArgs("new").WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "consumed_at"}).AddRow(1, now.Add(time.Minute), nil))
	mock.ExpectQuery("SELECT id FROM couples").WithArgs(int64(9), int64(1), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectExec("UPDATE moments SET").WithArgs(int64(9), string(domain.VisibilityCouple), "edited", string(domain.StatusMissYou), "外滩", nil, nil, moment.OccurredAt(), moment.UpdatedAt(), int64(20), int64(1), string(domain.SpaceShared), version).WillReturnResult(sqlmock.NewResult(0, 1))
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
	mock.ExpectQuery("SELECT id FROM couples").WithArgs(int64(9), int64(1), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
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

func TestListScopesQueriesToRequestedSpace(t *testing.T) {
	tests := []struct {
		name   string
		filter application.TimelineFilter
		query  string
		args   []any
	}{
		{
			name:   "shared relationship",
			filter: application.TimelineFilter{ViewerID: 1, RelationshipID: 9, SpaceType: domain.SpaceShared},
			query:  `FROM moments WHERE couple_id=\? AND space_type='shared' AND deleted_at IS NULL`,
			args:   []any{int64(9), 21},
		},
		{
			name:   "personal owner",
			filter: application.TimelineFilter{ViewerID: 1, OwnerID: 1, SpaceType: domain.SpacePersonal},
			query:  `FROM moments WHERE author_user_id=\? AND space_type='personal' AND deleted_at IS NULL`,
			args:   []any{int64(1), 21},
		},
		{
			name:   "partner visible personal",
			filter: application.TimelineFilter{ViewerID: 1, RelationshipID: 9, OwnerID: 2, SpaceType: domain.SpacePersonal},
			query:  `FROM moments WHERE couple_id=\? AND author_user_id=\? AND space_type='personal' AND visibility='partner' AND deleted_at IS NULL`,
			args:   []any{int64(9), int64(2), 21},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			expectation := mock.ExpectQuery(test.query)
			driverArgs := make([]driver.Value, len(test.args))
			for index, value := range test.args {
				driverArgs[index] = value
			}
			expectation.WithArgs(driverArgs...).WillReturnRows(sqlmock.NewRows([]string{"id"}))
			if _, err := NewMySQLRepository(db).List(context.Background(), test.filter, "", 20); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFindUsesViewerVisibilityScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(`FROM moments WHERE id=\? AND deleted_at IS NULL AND \(`).
		WithArgs(int64(7), int64(1), int64(9), int64(9), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	_, err = NewMySQLRepository(db).Find(context.Background(), application.MomentLookup{ViewerID: 1, RelationshipID: 9, MomentID: 7})
	if !errors.Is(err, domain.ErrMomentNotFound) {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSharedReadModelsExcludePersonalMoments(t *testing.T) {
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	t.Run("calendar", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("FROM moments WHERE couple_id=\\? AND space_type='shared' AND occurred_at").
			WithArgs(int64(9), from, to).
			WillReturnRows(sqlmock.NewRows([]string{"date", "count"}))
		if _, err := NewMySQLRepository(db).Calendar(context.Background(), 9, from, to); err != nil {
			t.Fatal(err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("map", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectQuery("FROM moments WHERE couple_id=\\? AND space_type='shared' AND occurred_at").
			WithArgs(int64(9), from, to, 500).
			WillReturnRows(sqlmock.NewRows([]string{"id", "latitude", "longitude", "occurred_at", "place_name"}))
		if _, err := NewMySQLRepository(db).Map(context.Background(), 9, from, to, 500); err != nil {
			t.Fatal(err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestShareToCoupleCopiesSourceAndPhotosAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	occurred := now.Add(-24 * time.Hour)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT author_user_id,space_type,kind,body,status_code,place_name,latitude,longitude,occurred_at FROM moments").WithArgs(int64(7), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"author_user_id", "space_type", "kind", "body", "status_code", "place_name", "latitude", "longitude", "occurred_at"}).AddRow(1, "personal", "standard", "原记录", "miss_you", "公园", nil, nil, occurred))
	mock.ExpectQuery("SELECT object_key,sort_order FROM moment_photos").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"object_key", "sort_order"}).AddRow("users/1/a.jpg", 1))
	mock.ExpectQuery("SELECT id FROM couples").WithArgs(int64(9), int64(1), int64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectExec("INSERT INTO moments").WithArgs(int64(9), int64(1), "shared", "couple", "standard", int64(7), nil, "也放进我们的故事", "miss_you", "公园", nil, nil, occurred, now, now).WillReturnResult(sqlmock.NewResult(20, 1))
	mock.ExpectExec("INSERT INTO moment_photos").WithArgs(int64(20), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	shared, err := NewMySQLRepository(db).ShareToCouple(context.Background(), 7, 1, 9, "也放进我们的故事", now)
	if err != nil {
		t.Fatal(err)
	}
	if shared.ID() != 20 || shared.SourceMomentID() == nil || *shared.SourceMomentID() != 7 || len(shared.Photos()) != 1 {
		t.Fatalf("shared=%+v photos=%+v", shared, shared.Photos())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReactionUpsertIsIdempotentAndDeleteIsOwnerScoped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	reaction, _ := domain.NewReaction(7, 2, domain.ReactionHug, "", now)
	mock.ExpectExec("INSERT INTO moment_reactions.*ON DUPLICATE KEY UPDATE").WithArgs(int64(7), int64(2), "hug", "", now, now).WillReturnResult(sqlmock.NewResult(12, 1))
	saved, err := NewMySQLRepository(db).UpsertReaction(context.Background(), reaction)
	if err != nil || saved.ID() != 12 {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
	mock.ExpectExec("DELETE FROM moment_reactions WHERE id=\\? AND user_id=\\?").WithArgs(int64(12), int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := NewMySQLRepository(db).DeleteReaction(context.Background(), 12, 2); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
