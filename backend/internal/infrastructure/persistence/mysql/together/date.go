package togethermysql

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// 纪念日的 anniversary_date 是 MySQL DATE 类型，只有一个"日历日"，没有时区概念。
// 但领域层用 time.Time 表示，前端传的是 `2026-10-15T00:00:00+08:00`。
// go-sql-driver 按 UTC 格式化 time.Time 时，`2026-10-15 00:00+08:00`
// 会变成 `2026-10-14 16:00 UTC`，存进 DATE 就成了 10-14 —— 比用户选的日期少一天。
//
// 修复方式：写入时显式转成东八区再取年月日；读取时把 DATE 当作"纯日期"解析，
// 保证「用户选了哪天，库里就是哪天，接口返回还是哪天」。
const anniversaryZone = 8 * 3600 // 固定 UTC+8：中国全境单一时区，无夏令时

var cst = time.FixedZone("CST", anniversaryZone)

// toAnniversaryDate 把领域层时间转成可直接写入 DATE 列的 `2006-01-02` 字符串。
func toAnniversaryDate(t time.Time) string {
	return t.In(cst).Format("2006-01-02")
}

// anniversaryDate 从数据库读取 DATE 列。
//
// DSN 里开了 parseTime=true 时驱动会直接把 DATE 转成 time.Time（UTC 零点）；
// 没开时则返回原始字节。两种都要支持，否则会因为格式不符直接报
// `parsing time "...": extra text`。
type anniversaryDate struct{ time.Time }

// Scan 实现 sql.Scanner。
func (d *anniversaryDate) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		return fmt.Errorf("anniversary_date 为空")
	case time.Time:
		// parseTime=true：驱动给的是 UTC 零点。转成东八区后仍是同一天，
		// 只取年月日，绝不能让时区运算把日期挪走。
		local := v.In(cst)
		d.Time = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, cst)
		return nil
	case []byte:
		return d.parse(string(v))
	case string:
		return d.parse(v)
	default:
		return fmt.Errorf("无法解析 anniversary_date: %T", src)
	}
}

func (d *anniversaryDate) parse(s string) error {
	if len(s) >= 10 {
		if parsed, err := time.ParseInLocation("2006-01-02", s[:10], cst); err == nil {
			d.Time = parsed
			return nil
		}
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("解析 anniversary_date %q 失败: %w", s, err)
	}
	local := parsed.In(cst)
	d.Time = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, cst)
	return nil
}

// Value 让 anniversaryDate 可以反向写入（测试与迁移场景）。
func (d anniversaryDate) Value() (driver.Value, error) { return d.Time, nil }
