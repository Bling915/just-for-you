package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/nanfang/just-for-you/server/internal/domain/identity"
)

func TestUpdateUserWithAvatarGrantConsumesGrantAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	user := domain.RestoreUser(7, "open", "小南", "users/7/avatars/new.jpg")
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id,purpose,expires_at,consumed_at FROM upload_grants").WithArgs("users/7/avatars/new.jpg").WillReturnRows(sqlmock.NewRows([]string{"user_id", "purpose", "expires_at", "consumed_at"}).AddRow(7, "avatar", now.Add(time.Minute), nil))
	mock.ExpectExec("UPDATE users SET display_name=.*,avatar_ref=").WithArgs("小南", "users/7/avatars/new.jpg", int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE upload_grants SET consumed_at=").WithArgs(now, "users/7/avatars/new.jpg").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err = (MySQLRepository{DB: db}).UpdateUserWithAvatarGrant(context.Background(), user, user.AvatarRef(), now); err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
