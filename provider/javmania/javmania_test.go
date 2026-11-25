package javmania

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestJAVMania_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"pacopacomama-101416_183",
		"caribbeancom-112426-001",
		"heyzo-3970",
		"av9898-2716",
		"jav-jusd-813",
	})
}

func TestJAVMania_GetMovieInfoByURL(t *testing.T) {
	testkit.Test(t, New, []string{
		"https://www.jav-mania.net/movies/1pondo/112326_001/",
		"https://www.jav-mania.net/fr/movies/caribbeanpremium/061826_001/",
		"https://www.jav-mania.net/jav/jusd-813",
	})
}

func TestJAVMania_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"paco 101416_183",
		"carib-112426-001",
		"HEYZO-3970",
		"JUSD-813",
	})
}

func TestJAVMania_GetActorInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"yui-hatano",
		"hibiki-otsuki",
	})
}

func TestJAVMania_SearchActor(t *testing.T) {
	testkit.Test(t, New, []string{
		"Yui Hatano",
		"大槻ひびき",
	})
}

func TestMovieCandidates(t *testing.T) {
	for keyword, want := range map[string][]string{
		"paco 101416_183.mp4":           {"pacopacomama-101416_183"},
		"[Thz.la]carib-112426-001.mp4":  {"caribbeancom-112426-001"},
		"carib-pr 112426_001":           {"caribbeanpremium-112426_001"},
		"carib.112426.001":              {"caribbeancom-112426-001"},
		"HEYZO-3970-FHD":                {"heyzo-3970"},
		"HEYZO-39701":                   nil,
		"HEYZO-39701-1080p.mp4":         nil,
		"112426_100-paco.mp4":           {"pacopacomama-112426_100"},
		"Caribbeancom 112426-001 x.mp4": {"caribbeancom-112426-001"},
		"112426-001.mp4":                {"caribbeancom-112426-001"},
		"112326_001.mp4":                {"1pondo-112326_001", "caribbeanpremium-112326_001", "pacopacomama-112326_001"},
		"10mu-112426_01":                {"10musume-112426_01"},
		"nyoshin n3424":                 {"nyoshin-n3424"},
		"h4610-ki261027":                {"h4610-ki261027"},
		"SSIS-001-C.mp4":                {"jav-ssis-001"},
		"ssis00001hhb.mp4":              {"jav-ssis-001"},
		"Movie2024.mp4":                 nil,
		"IMG-2024.mp4":                  nil,
		"FC2-PPV-1234567":               nil,
	} {
		assert.Equal(t, want, movieCandidates(keyword), keyword)
	}
}

func TestParseMovieIDFromURL(t *testing.T) {
	p := New()
	for u, want := range map[string]string{
		"https://www.jav-mania.net/movies/pacopacomama/101416_183/":    "pacopacomama-101416_183",
		"https://www.jav-mania.net/ko/movies/caribbeancom/112426-001/": "caribbeancom-112426-001",
		"https://www.jav-mania.net/jav/JUSD-813/":                      "jav-jusd-813",
	} {
		got, err := p.ParseMovieIDFromURL(u)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	}
	for _, bad := range []string{
		"https://www.jav-mania.net/actress/aika/",
		"https://www.jav-mania.net/movies/unknown/112326_001/",
		"https://www.jav-mania.net/movies/jav/abc-042/",
		"https://notjav-mania.net/jav/jusd-813/",
		"http://www.jav-mania.net/jav/jusd-813/",
	} {
		_, err := p.ParseMovieIDFromURL(bad)
		assert.Error(t, err, bad)
	}
}

func TestNormalizeMovieID(t *testing.T) {
	p := New()
	for in, want := range map[string]string{
		"pacopacomama-101416_183": "pacopacomama-101416_183",
		"JAV-SSIS-001":            "jav-ssis-001",
		"jav-a/b?c#d":             "",
		"caribbeancom-112426_001": "caribbeancom-112426-001",
		"paco 101416_183":         "pacopacomama-101416_183",
		"112326_001":              "",
		"ABC-00042":               "jav-abc-00042",
		"ABC-42":                  "jav-abc-042",
	} {
		assert.Equal(t, want, p.NormalizeMovieID(in), in)
	}
}
