package anniversarymysql

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/nanfang/just-for-you/server/internal/anniversary/domain"
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
	got, err := NewMySQLAnniversaryStore(db).Find(context.Background(), 9, 7)
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
	_, err := NewMySQLAnniversaryStore(db).Find(context.Background(), 9, 70)
	if errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-row database error must not become not found: %v", err)
	}

	db2, mock2, _ := sqlmock.New()
	defer db2.Close()
	mock2.ExpectQuery("SELECT .* FROM anniversaries").WithArgs(int64(9), int64(71)).WillReturnRows(sqlmock.NewRows([]string{"id", "couple_id", "creator_user_id", "title", "anniversary_date", "repeats_yearly", "count_mode", "pinned", "cover_ref", "message", "system_kind", "created_at", "updated_at"}))
	_, err = NewMySQLAnniversaryStore(db2).Find(context.Background(), 9, 71)
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
		WithArgs("旅行", "2026-10-22", false, "elapsed", true, "cover", "出发", now, int64(9), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := NewMySQLAnniversaryStore(db).Update(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// 防回归：领域层传进来的是东八区零点（如前端 `2026-10-15T00:00:00+08:00`），
// 写入 DATE 列时必须原样保留年月日，不能因转 UTC 而少一天。
func TestAnniversaryDateKeepsEast8CalendarDay(t *testing.T) {
	east8 := time.FixedZone("CST", 8*3600)
	cases := []struct {
		name string
		in   time.Time
		want string
	}{
		{"东八区零点", time.Date(2026, 10, 15, 0, 0, 0, 0, east8), "2026-10-15"},
		{"东八区当天傍晚", time.Date(2026, 10, 15, 23, 30, 0, 0, east8), "2026-10-15"},
		{"UTC 零点（已是同一日历日）", time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), "2026-10-15"},
		{"跨月边界", time.Date(2026, 3, 1, 0, 0, 0, 0, east8), "2026-03-01"},
		{"跨年边界", time.Date(2027, 1, 1, 0, 0, 0, 0, east8), "2027-01-01"},
		{"闰日", time.Date(2028, 2, 29, 0, 0, 0, 0, east8), "2028-02-29"},
	}
	for _, tc := range cases {
		if got := toAnniversaryDate(tc.in); got != tc.want {
			t.Errorf("%s: toAnniversaryDate(%v) = %q，期望 %q", tc.name, tc.in, got, tc.want)
		}
	}
}

// 防回归：读取 DATE 列时也必须还原成东八区当天零点，
// 且要兼容 DSN 开不开 parseTime=true 的两种返回形态。
func TestAnniversaryDateScanRoundTrip(t *testing.T) {
	east8 := time.FixedZone("CST", 8*3600)
	want := time.Date(2026, 10, 15, 0, 0, 0, 0, east8)

	// parseTime=true：驱动返回 UTC 零点
	var fromTime anniversaryDate
	if err := fromTime.Scan(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("扫描 time.Time 失败: %v", err)
	}
	if !fromTime.Time.Equal(want) {
		t.Errorf("扫描 time.Time 得 %v，期望 %v", fromTime.Time, want)
	}

	// parseTime=false：驱动返回原始字节
	for _, raw := range []any{[]byte("2026-10-15"), "2026-10-15"} {
		var fromRaw anniversaryDate
		if err := fromRaw.Scan(raw); err != nil {
			t.Fatalf("扫描 %T 失败: %v", raw, err)
		}
		if !fromRaw.Time.Equal(want) {
			t.Errorf("扫描 %T 得 %v，期望 %v", raw, fromRaw.Time, want)
		}
	}

	// 写→读 往返一致
	stored := toAnniversaryDate(want)
	var back anniversaryDate
	if err := back.Scan([]byte(stored)); err != nil {
		t.Fatalf("往返扫描失败: %v", err)
	}
	if !back.Time.Equal(want) {
		t.Errorf("往返不一致: %v → %q → %v", want, stored, back.Time)
	}
}
