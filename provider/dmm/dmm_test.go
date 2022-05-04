package dmm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDMM_ParseNumber(t *testing.T) {
	dmm := &DMM{}
	for _, unit := range []struct {
		id, want string
	}{
		{"ssis00123", "SSIS-123"},
		{"48midv00123", "MIDV-123"},
		{"48midv00003", "MIDV-003"},
		{"24ped00020", "PED-020"},
		{"abc00120", "ABC-120"},
		{"abc00120l", "ABC-120"},
		{"19abc00120l", "ABC-120"},
		{"abc00001", "ABC-001"},
		{"h_001fcp00006", "FCP-006"},
		{"001fcp06", "FCP-006"},
		{"h_001fcp06", "FCP-006"},
		{"scute1192", "SCUTE-1192"},
		{"h_198need00094r18", "NEED-094"},
		{"1fsdss00131re01", "FSDSS-131"},
	} {
		assert.Equal(t, unit.want, dmm.ParseNumber(unit.id))
	}
}

func TestDMM_PreviewSrc(t *testing.T) {
	dmm := &DMM{}
	for _, unit := range []struct {
		src, want string
	}{
		{"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990ps.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990pl.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/consumer_game/pppd00990/pppd00990js-1.jpg",
			"https://pics.dmm.co.jp/digital/consumer_game/pppd00990/pppd00990-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990js-1.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990jp-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990ts-1.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990tl-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990-1.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990jp-1.jpg",
		},
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990-23.jpg",
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990jp-23.jpg",
		},
	} {
		assert.Equal(t, unit.want, dmm.PreviewSrc(unit.src))
	}
}
