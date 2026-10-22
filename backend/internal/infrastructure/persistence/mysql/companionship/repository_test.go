package infrastructure

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/companionship"
)

func TestRepositoryStoresPresenceAndReadsLatestEffectiveCandidate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	expires := now.Add(time.Hour)
	presence, _ := domain.NewPresence(9, 7, domain.PresetInClass, "", &expires, now)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO presences(couple_id,user_id,preset,custom_text,expires_at,updated_at) VALUES(?,?,?,?,?,?)")).WithArgs(int64(9), int64(7), domain.PresetInClass, "", &expires, now).WillReturnResult(sqlmock.NewResult(12, 1))
	stored, err := repo.SavePresence(context.Background(), presence)
	if err != nil || stored.ID() != 12 {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	mock.ExpectQuery("SELECT id,couple_id,user_id,preset,custom_text,expires_at,updated_at FROM presences").WithArgs(int64(9), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "couple_id", "user_id", "preset", "custom_text", "expires_at", "updated_at"}).AddRow(12, 9, 7, "in_class", "", expires, now))
	latest, err := repo.LatestPresence(context.Background(), 9, 7)
	if err != nil || latest == nil || latest.ID() != 12 {
		t.Fatalf("latest=%+v err=%v", latest, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryReturnsOnlyActiveRelationship(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewRepository(db)
	mock.ExpectQuery("SELECT id,user_a_id,user_b_id FROM couples").WithArgs(int64(7), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id", "user_a_id", "user_b_id"}).AddRow(9, 7, 8))
	relationship, err := repo.GetActiveRelationship(context.Background(), 7)
	if err != nil || relationship.ID != 9 || relationship.UserBID != 8 {
		t.Fatalf("relationship=%+v err=%v", relationship, err)
	}
}
