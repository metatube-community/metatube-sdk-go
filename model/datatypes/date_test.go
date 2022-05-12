package datatypes

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate_MarshalJSON(t *testing.T) {
	for _, unit := range []struct {
		t    time.Time
		want string
	}{
		{time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), `"2000-01-01"`},
		{time.Date(2000, 1, 1, 1, 0, 0, 0, time.UTC), `"2000-01-01"`},
		{time.Date(2000, 1, 1, 0, 1, 0, 0, time.UTC), `"2000-01-01"`},
		{time.Date(2000, 1, 1, 0, 0, 1, 0, time.UTC), `"2000-01-01"`},
		{time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC), `"2000-02-29"`},
	} {
		data, _ := json.Marshal(Date(unit.t))
		assert.Equal(t, unit.want, string(data))
	}
}

func TestDate_UnmarshalJSON(t *testing.T) {
	for _, unit := range []struct {
		want time.Time
		d    string
	}{
		{time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), `"2000-01-01"`},
		{time.Date(1991, 12, 14, 0, 0, 0, 0, time.UTC), `"1991-12-14"`},
		{time.Date(2000, 8, 8, 0, 0, 0, 0, time.UTC), `"2000-08-08"`},
		{time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC), `"2000-02-29"`},
	} {
		date := Date{}
		_ = json.Unmarshal([]byte(unit.d), &date)
		assert.True(t, unit.want.Equal(time.Time(date)))
	}
}
