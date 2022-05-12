package datatypes

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"
)

const dateFormat = "2006-01-02"

type Date time.Time

func (date *Date) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	*date = Date(nullTime.Time)
	return
}

func (date Date) Value() (driver.Value, error) {
	y, m, d := time.Time(date).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Time(date).Location()), nil
}

func (date Date) String() string {
	return time.Time(date).Format(dateFormat)
}

// GormDataType gorm common data type
func (date Date) GormDataType() string {
	return "date"
}

func (date Date) GobEncode() ([]byte, error) {
	return time.Time(date).GobEncode()
}

func (date *Date) GobDecode(b []byte) error {
	return (*time.Time)(date).GobDecode(b)
}

func (date Date) MarshalJSON() ([]byte, error) {
	t := time.Time(date)
	if y := t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	t.String()

	b := make([]byte, 0, len(dateFormat)+2)
	b = append(b, '"')
	b = t.AppendFormat(b, dateFormat)
	b = append(b, '"')
	return b, nil
}

func (date *Date) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	t, err := time.Parse(`"`+dateFormat+`"`, string(data))
	if err != nil {
		return err
	}
	*date = Date(t)
	return nil
}
