package javdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/metatube-community/metatube-sdk-go/provider"
)

type testConfig map[string]string

func (c testConfig) Has(key string) bool { _, ok := c[key]; return ok }
func (c testConfig) GetString(key string) (string, error) {
	return c[key], nil
}
func (testConfig) GetBool(string) (bool, error)              { return false, nil }
func (testConfig) GetInt64(string) (int64, error)            { return 0, nil }
func (testConfig) GetFloat64(string) (float64, error)        { return 0, nil }
func (testConfig) GetDuration(string) (time.Duration, error) { return 0, nil }

func TestSetConfig(t *testing.T) {
	j := New()
	require.Zero(t, j.Priority())
	require.NoError(t, j.SetConfig(testConfig{}))
	require.Zero(t, j.Priority())

	require.NoError(t, j.SetConfig(testConfig{"cookie": "session=value"}))
	require.EqualValues(t, Priority, j.Priority())
	require.Equal(t, "session=value", j.cookie)
	require.Equal(t, defaultUserAgent, j.userAgent)

	require.NoError(t, j.SetConfig(testConfig{"cookie": "session=value", "user_agent": "browser"}))
	require.Equal(t, "browser", j.userAgent)

	require.NoError(t, j.SetConfig(testConfig{"cookie": "  "}))
	require.Zero(t, j.Priority())
	require.Empty(t, j.cookie)
}

func TestParseSearchResults(t *testing.T) {
	j := New()
	results, err := j.parseSearchResults([]byte(`
		<div class="item">
		  <a class="box" href="/v/wKg2O2"></a>
		  <div class="video-title"><strong>HEYZO-3839</strong> Example title</div>
		  <div class="meta">2026-08-01</div>
		  <div class="score">4.2分</div>
		  <div class="cover"><img src="/covers/example.jpg"></div>
		</div>`))
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "wKg2O2", results[0].ID)
	require.Equal(t, "HEYZO-3839", results[0].Number)
	require.Equal(t, "Example title", results[0].Title)
	require.Equal(t, "https://javdb.com/v/wKg2O2", results[0].Homepage)
	require.Equal(t, "https://javdb.com/covers/example.jpg", results[0].CoverURL)
	require.Equal(t, 4.2, results[0].Score)
	require.True(t, results[0].IsValid())
}

func TestParseSearchResultsEmptyAndChallenge(t *testing.T) {
	j := New()
	_, err := j.parseSearchResults([]byte(`<div class="empty-message">沒有資料</div>`))
	require.ErrorIs(t, err, provider.ErrInfoNotFound)

	_, err = j.parseSearchResults([]byte(`<script src="/cdn-cgi/challenge-platform/h/g/orchestrate/chl_page/v1"></script>`))
	require.ErrorContains(t, err, "anti-bot challenge")
}

func TestParseMovieInfo(t *testing.T) {
	j := New()
	info, err := j.parseMovieInfo([]byte(`
		<h2 class="title is-4 current-title">Example title</h2>
		<img class="video-cover" src="/covers/example.jpg">
		<div class="movie-panel-info">
		  <div class="panel-block"><strong>番號:</strong><span class="value"><span data-clipboard-text="HEYZO-3839"></span></span></div>
		  <div class="panel-block"><strong>日期:</strong><span class="value">2026-08-01</span></div>
		  <div class="panel-block"><strong>時長:</strong><span class="value">120 分鐘</span></div>
		  <div class="panel-block"><strong>評分:</strong><span class="value">4.2分</span></div>
		  <div class="panel-block"><strong>導演:</strong><span class="value"><a>Director</a></span></div>
		  <div class="panel-block"><strong>片商:</strong><span class="value"><a>Maker</a></span></div>
		  <div class="panel-block"><strong>系列:</strong><span class="value"><a>Series</a></span></div>
		  <div class="panel-block"><strong>類別:</strong><span class="value"><a>Genre A</a><a>Genre B</a></span></div>
		  <div class="panel-block"><strong>演員:</strong><span class="value"><a>Actor A</a></span></div>
		</div>
		<div class="preview-images"><a class="tile-item" href="/samples/1.jpg"><img></a></div>`), "wKg2O2", "https://javdb.com/v/wKg2O2")
	require.NoError(t, err)
	require.Equal(t, "HEYZO-3839", info.Number)
	require.Equal(t, "Example title", info.Title)
	require.Equal(t, "https://javdb.com/covers/example.jpg", info.CoverURL)
	require.Equal(t, 120, info.Runtime)
	require.Equal(t, "Director", info.Director)
	require.Equal(t, "Maker", info.Maker)
	require.Equal(t, "Series", info.Series)
	require.Equal(t, []string{"Actor A"}, []string(info.Actors))
	require.Equal(t, []string{"Genre A", "Genre B"}, []string(info.Genres))
	require.Equal(t, []string{"https://javdb.com/samples/1.jpg"}, []string(info.PreviewImages))
	require.True(t, info.IsValid())
}

func TestNormalizeAndParseMovieID(t *testing.T) {
	j := New()
	require.Equal(t, "HEYZO-3839", j.NormalizeMovieKeyword("HEYZO-3839-C.mkv"))
	require.Equal(t, "wKg2O2", j.NormalizeMovieID("wKg2O2"))
	require.Empty(t, j.NormalizeMovieID("../wKg2O2"))

	id, err := j.ParseMovieIDFromURL("https://javdb.com/v/wKg2O2?locale=zh")
	require.NoError(t, err)
	require.Equal(t, "wKg2O2", id)
	_, err = j.ParseMovieIDFromURL("https://example.com/v/wKg2O2")
	require.ErrorIs(t, err, provider.ErrInvalidURL)
}
