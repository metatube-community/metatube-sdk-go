package parser

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
		{"apx.1min", time.Minute},
		{"about 1min", time.Minute},
		{"1分", time.Minute},
		{"1分1秒", time.Minute + time.Second},
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

func TestParseActorNames(t *testing.T) {
	for _, unit := range []struct {
		orig string
		want []string
	}{
		{"  ", nil},
		{"川上ゆう", []string{"川上ゆう"}},
		{"川上ゆう 20歲", []string{"川上ゆう 20歲"}},
		{"（森野雫）", []string{"森野雫"}},
		{"川上ゆう（森野雫）", []string{"川上ゆう", "森野雫"}},
		{"新井エリー（晶エリー、大沢佑香）", []string{"新井エリー", "晶エリー", "大沢佑香"}},
	} {
		assert.ElementsMatch(t, unit.want, ParseActorNames(unit.orig), fmt.Sprintf("Arg: %s", unit.orig))
	}
}

func TestParseIDToNumber(t *testing.T) {
	for _, unit := range []struct {
		id, want string
	}{
		{"mdx0109", "MDX-0109"},
		{"mdx-0264", "MDX-0264"},
		{"91cm109", "91CM-109"},
		{"91CM-109", "91CM-109"},
		{"dldss287", "DLDSS-287"},
	} {
		assert.Equal(t, unit.want, ParseIDToNumber(unit.id))
	}
}
