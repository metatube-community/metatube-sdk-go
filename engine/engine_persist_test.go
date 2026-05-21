package engine

import (
	"net/url"
	"testing"
	"time"

	"github.com/lib/pq"
	"golang.org/x/text/language"
	"gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/database"
	mt "github.com/metatube-community/metatube-sdk-go/provider"

	"github.com/metatube-community/metatube-sdk-go/model"
)

// newTestEngine builds an Engine backed by a fresh in-memory SQLite, with all
// model tables migrated. Each call returns an isolated DB.
func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	// Use a unique DSN per test so parallel runs don't share state. SQLite's
	// `:memory:` is per-connection; `file::memory:?cache=shared` is shared
	// across connections of the SAME DSN, so we vary the name.
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := database.Open(&database.Config{DSN: dsn, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}
	e := New(db)
	if err := e.DBAutoMigrate(true); err != nil {
		t.Fatalf("DBAutoMigrate: %v", err)
	}
	return e
}

// fakeMovieProvider implements mt.MovieProvider for tests. Its GetMovieInfoByID
// returns whatever the test sets in `next`.
type fakeMovieProvider struct {
	name string
	url  *url.URL
	next *model.MovieInfo
	err  error
}

func (f *fakeMovieProvider) Name() string                            { return f.name }
func (f *fakeMovieProvider) Priority() float64                       { return 1 }
func (f *fakeMovieProvider) SetPriority(v float64)                   {}
func (f *fakeMovieProvider) Language() language.Tag                  { return language.Japanese }
func (f *fakeMovieProvider) URL() *url.URL                           { return f.url }
func (f *fakeMovieProvider) NormalizeMovieID(id string) string       { return id }
func (f *fakeMovieProvider) ParseMovieIDFromURL(string) (string, error) {
	return "", nil
}
func (f *fakeMovieProvider) GetMovieInfoByID(id string) (*model.MovieInfo, error) {
	return f.next, f.err
}
func (f *fakeMovieProvider) GetMovieInfoByURL(string) (*model.MovieInfo, error) {
	return f.next, f.err
}

var _ mt.MovieProvider = (*fakeMovieProvider)(nil)

func newFakeMovieProvider(name string) *fakeMovieProvider {
	u, _ := url.Parse("https://example.com/")
	return &fakeMovieProvider{name: name, url: u}
}

func fullMovieInfo(id, provider string) *model.MovieInfo {
	return &model.MovieInfo{
		ID:                 id,
		Number:             "NUM-001",
		Title:              "Full Title",
		Summary:            "Full summary, the good one.",
		Provider:           provider,
		Homepage:           "https://example.com/" + id,
		Director:           "Director X",
		Actors:             pq.StringArray{"actor1", "actor2"},
		ThumbURL:           "https://example.com/thumb.jpg",
		BigThumbURL:        "https://example.com/big-thumb.jpg",
		CoverURL:           "https://example.com/cover.jpg",
		BigCoverURL:        "https://example.com/big-cover.jpg",
		PreviewVideoURL:    "https://example.com/preview.mp4",
		PreviewVideoHLSURL: "https://example.com/preview.m3u8",
		PreviewImages:      pq.StringArray{"https://example.com/p1.jpg"},
		Maker:              "Maker M",
		Label:              "Label L",
		Series:             "Series S",
		Genres:             pq.StringArray{"genre1", "genre2"},
		Score:              4.5,
		Runtime:            120,
		ReleaseDate:        datatypes.Date(time.Date(2024, 2, 13, 0, 0, 0, 0, time.UTC)),
	}
}

// minimalMovieInfo returns a MovieInfo that passes IsValid() but has every
// optional field at its zero value - the "regression case" that used to
// clobber the DB.
func minimalMovieInfo(id, provider string) *model.MovieInfo {
	return &model.MovieInfo{
		ID:       id,
		Number:   "NUM-001",
		Title:    "Full Title",
		CoverURL: "https://example.com/cover.jpg",
		Provider: provider,
		Homepage: "https://example.com/" + id,
	}
}

// TestGetMovieInfo_EmptyFieldsDoNotClobberExisting is the central regression
// test for the JAV321/JavBus "empty summary overwrites good summary" bug. It
// reproduces the failure sequence:
//
//  1. Fetch #1 returns a fully-populated MovieInfo. DB persists it.
//  2. Fetch #2 (e.g. provider had a transient miss, or it's a provider like
//     JavBus that doesn't populate Summary) returns a MovieInfo with the
//     required IsValid() fields but every optional field empty.
//
// With UpdateAll: true and no PreserveFrom, step 2 used to wipe Summary,
// Director, Actors, Maker, etc. from the DB row. With PreserveFrom, the row
// retains its previously-stored values.
func TestGetMovieInfo_EmptyFieldsDoNotClobberExisting(t *testing.T) {
	e := newTestEngine(t)
	p := newFakeMovieProvider("FakeProvider")

	id := "sone00052"

	// --- Step 1: fully-populated fetch.
	full := fullMovieInfo(id, p.name)
	p.next = full
	got1, err := e.getMovieInfoByProviderID(p, id, false /* lazy */)
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}
	if got1.Summary != full.Summary {
		t.Fatalf("step 1: returned info should have full Summary, got %q", got1.Summary)
	}

	// Verify DB now has the full info.
	stored, err := e.getMovieInfoFromDB(p, id)
	if err != nil {
		t.Fatalf("DB read after step 1: %v", err)
	}
	if stored.Summary != full.Summary {
		t.Fatalf("step 1: DB Summary mismatch, got %q want %q", stored.Summary, full.Summary)
	}
	if len(stored.Actors) != 2 {
		t.Fatalf("step 1: DB Actors mismatch, got %v", stored.Actors)
	}

	// --- Step 2: provider returns minimal info (the bug trigger).
	p.next = minimalMovieInfo(id, p.name)
	got2, err := e.getMovieInfoByProviderID(p, id, false /* lazy */)
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}

	// The returned struct should also be enriched - PreserveFrom mutates the
	// info before upsert, so callers see the merged result.
	if got2.Summary != full.Summary {
		t.Errorf("step 2: returned Summary should be preserved, got %q want %q",
			got2.Summary, full.Summary)
	}

	// Re-read from DB to confirm persistence.
	stored2, err := e.getMovieInfoFromDB(p, id)
	if err != nil {
		t.Fatalf("DB read after step 2: %v", err)
	}
	if stored2.Summary != full.Summary {
		t.Errorf("step 2: DB Summary was clobbered, got %q want %q",
			stored2.Summary, full.Summary)
	}
	if stored2.Director != full.Director {
		t.Errorf("step 2: DB Director was clobbered, got %q want %q",
			stored2.Director, full.Director)
	}
	if len(stored2.Actors) != len(full.Actors) {
		t.Errorf("step 2: DB Actors was clobbered, got %v want %v",
			stored2.Actors, full.Actors)
	}
	if len(stored2.Genres) != len(full.Genres) {
		t.Errorf("step 2: DB Genres was clobbered, got %v want %v",
			stored2.Genres, full.Genres)
	}
	if stored2.Score != full.Score {
		t.Errorf("step 2: DB Score was clobbered, got %v want %v", stored2.Score, full.Score)
	}
	if stored2.Maker != full.Maker {
		t.Errorf("step 2: DB Maker was clobbered, got %q want %q", stored2.Maker, full.Maker)
	}
	if stored2.Label != full.Label {
		t.Errorf("step 2: DB Label was clobbered, got %q want %q", stored2.Label, full.Label)
	}
	if stored2.Series != full.Series {
		t.Errorf("step 2: DB Series was clobbered, got %q want %q", stored2.Series, full.Series)
	}
}

// TestGetMovieInfo_NonEmptyFieldsWinOverExisting verifies the converse: when
// the provider DOES return a non-empty value, it should overwrite the DB. We
// must not regress into a "DB always wins" model where providers can never
// update their own data.
func TestGetMovieInfo_NonEmptyFieldsWinOverExisting(t *testing.T) {
	e := newTestEngine(t)
	p := newFakeMovieProvider("FakeProvider")

	id := "id-1"

	// Seed DB with a v1 of the movie.
	p.next = fullMovieInfo(id, p.name)
	if _, err := e.getMovieInfoByProviderID(p, id, false); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Provider now returns a v2 with updated Summary, Score, and an extra actor.
	v2 := fullMovieInfo(id, p.name)
	v2.Summary = "Updated summary - v2."
	v2.Score = 4.9
	v2.Actors = pq.StringArray{"new-actor"}
	v2.Genres = pq.StringArray{"new-genre"}
	p.next = v2
	if _, err := e.getMovieInfoByProviderID(p, id, false); err != nil {
		t.Fatalf("v2 fetch: %v", err)
	}

	stored, err := e.getMovieInfoFromDB(p, id)
	if err != nil {
		t.Fatalf("DB read: %v", err)
	}
	if stored.Summary != "Updated summary - v2." {
		t.Errorf("Summary v2 should win, got %q", stored.Summary)
	}
	if stored.Score != 4.9 {
		t.Errorf("Score v2 should win, got %v", stored.Score)
	}
	// Actors should be replaced (not appended).
	if len(stored.Actors) != 1 || stored.Actors[0] != "new-actor" {
		t.Errorf("Actors v2 should replace v1, got %v", stored.Actors)
	}
	if len(stored.Genres) != 1 || stored.Genres[0] != "new-genre" {
		t.Errorf("Genres v2 should replace v1, got %v", stored.Genres)
	}
}

// TestGetMovieInfo_LazyHitDoesNotTriggerUpsert verifies that the lazy path
// (DB hit on a valid record) returns early WITHOUT touching the DB. This is
// important because the auto-save defer is registered AFTER the lazy check;
// our PreserveFrom logic lives inside that defer and must not fire on lazy
// DB hits.
func TestGetMovieInfo_LazyHitDoesNotTriggerUpsert(t *testing.T) {
	e := newTestEngine(t)
	p := newFakeMovieProvider("FakeProvider")

	id := "id-1"

	// Seed.
	p.next = fullMovieInfo(id, p.name)
	if _, err := e.getMovieInfoByProviderID(p, id, false); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Configure the provider to fail loudly if called - lazy hit should
	// short-circuit and never invoke the provider callback.
	called := false
	p.next = nil
	p.err = nil
	wrappedProvider := &countingProvider{fakeMovieProvider: p, callCount: &called}

	got, err := e.getMovieInfoByProviderID(wrappedProvider, id, true /* lazy */)
	if err != nil {
		t.Fatalf("lazy fetch: %v", err)
	}
	if called {
		t.Errorf("lazy hit should not invoke provider, but GetMovieInfoByID was called")
	}
	if got.Summary != "Full summary, the good one." {
		t.Errorf("lazy hit should return DB record, got Summary=%q", got.Summary)
	}
}

// countingProvider wraps fakeMovieProvider and flips a flag when
// GetMovieInfoByID is called.
type countingProvider struct {
	*fakeMovieProvider
	callCount *bool
}

func (c *countingProvider) GetMovieInfoByID(id string) (*model.MovieInfo, error) {
	*c.callCount = true
	return c.fakeMovieProvider.GetMovieInfoByID(id)
}

// TestGetMovieInfo_FreshInsertWorks verifies the baseline: a first-time fetch
// with no existing DB record stores the full info verbatim. PreserveFrom must
// be a no-op when there's nothing to preserve from.
func TestGetMovieInfo_FreshInsertWorks(t *testing.T) {
	e := newTestEngine(t)
	p := newFakeMovieProvider("FakeProvider")

	id := "brand-new"
	p.next = fullMovieInfo(id, p.name)

	got, err := e.getMovieInfoByProviderID(p, id, false)
	if err != nil {
		t.Fatalf("fresh insert: %v", err)
	}
	if got.Summary != p.next.Summary {
		t.Errorf("Summary mismatch, got %q want %q", got.Summary, p.next.Summary)
	}

	stored, err := e.getMovieInfoFromDB(p, id)
	if err != nil {
		t.Fatalf("DB read: %v", err)
	}
	if stored.Summary != p.next.Summary {
		t.Errorf("DB Summary mismatch, got %q want %q", stored.Summary, p.next.Summary)
	}
	if len(stored.Actors) != 2 {
		t.Errorf("DB Actors mismatch, got %v", stored.Actors)
	}
}
