package fanza

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFANZA_GetMovieInfoByID(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"1silk00113",
		"adn00306",
		"1sdjs00033",
		"okax841",
		"zuko00122",
		"118chn064",
		"h_346rebd655tk2",
		"midv00047",
		"403jdxa57676",
		"pkpd00170",
		"mism00238",
		"1msfh00007",
		"1stars00141",
		"vrkm00722",
	} {
		info, err := provider.GetMovieInfoByID(item)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.True(t, assert.NoError(t, err) && assert.True(t, info.Valid()))
		t.Logf("%s", data)
	}
}

func TestFANZA_SearchMovie(t *testing.T) {
	provider := New()
	for _, item := range []string{
		"OREC-062",
		"T-28621",
		"midv-003",
		"blk-052",
		"mogi044",
		"SSNI-379",
		"SSIS-122",
		"MIDV-047",
		"abw",
	} {
		results, err := provider.SearchMovie(provider.NormalizeKeyword(item))
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}

func TestParseNumber(t *testing.T) {
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
		{"h_068mxgs1184bod", "MXGS-1184"},
		{"h_093r1800258", "R-1800258"},
		{"55t28621tk", "T-28621"},
	} {
		assert.Equal(t, unit.want, ParseNumber(unit.id))
	}
}

func TestPreviewSrc(t *testing.T) {
	for _, unit := range []struct {
		src, want string
	}{
		{
			"https://pics.dmm.co.jp/digital/video/pppd00990/pppd00990ps.jpg",
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
		assert.Equal(t, unit.want, PreviewSrc(unit.src))
	}
}
