package providerid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	for _, unit := range []struct {
		pid, want string
	}{
		{"FANZA:mdx0109", "FANZA:mdx0109"},
		{"FANZA:mdx0109:0.9", "FANZA:mdx0109"},
		{"FANZA:mdx0109:0", "FANZA:mdx0109"},
		{"FANZA:mdx0109:1", "FANZA:mdx0109"},
		{"FANZA:mdx0109:1.2", "FANZA:mdx0109"},
		{"AVBASE:dmm:ssis899", "AVBASE:dmm:ssis899"},
		{"AVBASE:dmm%3Assis899", "AVBASE:dmm:ssis899"},
		{"AVBASE:dmm%3assis899", "AVBASE:dmm:ssis899"},
		{"AVBASE:dmm:ssis899:0.99", "AVBASE:dmm:ssis899"},
		{"ARZON:2234", "ARZON:2234"},
		{"ARZON:2234:0.55", "ARZON:2234"},
		{"ARZON:2234:1233", "ARZON:2234:1233"},
	} {
		pid, err := Parse(unit.pid)
		if assert.NoError(t, err) {
			got := pid.Provider + ":" + pid.ID
			assert.Equal(t, unit.want, got)
		}
	}
}
