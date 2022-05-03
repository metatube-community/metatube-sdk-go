package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	for _, unit := range []struct {
		orig string
		want time.Duration
	}{
		{"0", 0},
		{"1min", time.Minute},
		{"1分", time.Minute},
		{"1h2m3s", time.Hour + time.Minute*2 + time.Second*3},
		{"1H2M3S", time.Hour + time.Minute*2 + time.Second*3},
		{"PT1H2M3S", time.Hour + time.Minute*2 + time.Second*3},
		{"t01h02m03s", time.Hour + time.Minute*2 + time.Second*3},
		{"03s", time.Second * 3},
		{"02m03s", time.Minute*2 + time.Second*3},
		{"pt02m03s", time.Minute*2 + time.Second*3},
		{"01:02:03", time.Hour + time.Minute*2 + time.Second*3},
		{"PT1:2:03", time.Hour + time.Minute*2 + time.Second*3},
		{"PT01:02:03", time.Hour + time.Minute*2 + time.Second*3},
	} {
		assert.Equal(t, unit.want, ParseDuration(unit.orig), fmt.Sprintf("Arg: %s", unit.orig))
	}
}
