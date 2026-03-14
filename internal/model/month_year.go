package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const LayoutMMYYYY = "01-2006"

type MonthYear time.Time

func (my *MonthYear) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse(LayoutMMYYYY, s)
	if err != nil {
		return fmt.Errorf("invalid date format: expected MM-YYYY, got %s", s)
	}
	*my = MonthYear(t)
	return nil
}

func (my MonthYear) MarshalJSON() ([]byte, error) {
	t := time.Time(my)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, t.Format(LayoutMMYYYY))), nil
}

func (my MonthYear) Time() time.Time {
	return time.Time(my)
}

// Scan implements the sql.Scanner interface.
func (my *MonthYear) Scan(value interface{}) error {
	if value == nil {
		*my = MonthYear(time.Time{})
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("cannot scan type %T into MonthYear", value)
	}
	*my = MonthYear(t)
	return nil
}

// Value implements the driver.Valuer interface.
func (my MonthYear) Value() (driver.Value, error) {
	t := time.Time(my)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}
