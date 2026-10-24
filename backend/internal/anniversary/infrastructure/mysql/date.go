package anniversarymysql

import (
	"database/sql/driver"
	"fmt"
	"time"
)

var cst = time.FixedZone("CST", 8*3600)

func toAnniversaryDate(value time.Time) string {
	return value.In(cst).Format("2006-01-02")
}

type anniversaryDate struct {
	time.Time
}

func (date *anniversaryDate) Scan(source any) error {
	switch value := source.(type) {
	case time.Time:
		local := value.In(cst)
		date.Time = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, cst)
		return nil
	case string:
		return date.parse(value)
	case []byte:
		return date.parse(string(value))
	default:
		return fmt.Errorf("unsupported anniversary date type %T", source)
	}
}

func (date *anniversaryDate) parse(value string) error {
	parsed, err := time.ParseInLocation("2006-01-02", value[:min(10, len(value))], cst)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		local := parsed.In(cst)
		parsed = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, cst)
	}
	date.Time = parsed
	return nil
}

func (date anniversaryDate) Value() (driver.Value, error) {
	return date.Time, nil
}
