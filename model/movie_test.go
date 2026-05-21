package model

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

func TestMovieInfo_IsValid(t *testing.T) {
	// Lock current IsValid semantics: Summary is NOT required.
	// PreserveFrom relies on this invariant - changing IsValid to require
	// Summary would break providers (e.g. JavBus) that never set it.
	t.Parallel()

	base := MovieInfo{
		ID:       "abc",
		Number:   "ABC-001",
		Title:    "Title",
		CoverURL: "https://example.com/cover.jpg",
		Provider: "Test",
		Homepage: "https://example.com/abc",
	}
	if !base.IsValid() {
		t.Fatalf("baseline MovieInfo should be valid, got invalid: %+v", base)
	}

	withoutSummary := base
	withoutSummary.Summary = ""
	if !withoutSummary.IsValid() {
		t.Errorf("MovieInfo with empty Summary should still be valid (Summary is optional)")
	}

	for _, tc := range []struct {
		name string
		mut  func(*MovieInfo)
	}{
		{"missing ID", func(m *MovieInfo) { m.ID = "" }},
		{"missing Number", func(m *MovieInfo) { m.Number = "" }},
		{"missing Title", func(m *MovieInfo) { m.Title = "" }},
		{"missing CoverURL", func(m *MovieInfo) { m.CoverURL = "" }},
		{"missing Provider", func(m *MovieInfo) { m.Provider = "" }},
		{"missing Homepage", func(m *MovieInfo) { m.Homepage = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base
			tc.mut(&m)
			if m.IsValid() {
				t.Errorf("MovieInfo missing required field should be invalid: %+v", m)
			}
		})
	}
}

func TestMovieInfo_PreserveFrom_NilExisting(t *testing.T) {
	t.Parallel()
	m := &MovieInfo{ID: "abc"}
	// Must not panic, must not mutate.
	m.PreserveFrom(nil)
	if m.ID != "abc" {
		t.Errorf("PreserveFrom(nil) should not touch any field, got ID=%q", m.ID)
	}
}

// fullExisting returns an existing MovieInfo with every optional field populated.
// Used as the "previously stored good data" baseline for clobber-prevention tests.
func fullExisting() *MovieInfo {
	return &MovieInfo{
		ID:                 "id-1",
		Number:             "NUM-001",
		Title:              "Existing Title",
		Summary:            "Existing summary, the good one.",
		Provider:           "TestProvider",
		Homepage:           "https://example.com/id-1",
		Director:           "Existing Director",
		Actors:             pq.StringArray{"actor1", "actor2"},
		ThumbURL:           "https://example.com/thumb.jpg",
		BigThumbURL:        "https://example.com/big-thumb.jpg",
		CoverURL:           "https://example.com/cover.jpg",
		BigCoverURL:        "https://example.com/big-cover.jpg",
		PreviewVideoURL:    "https://example.com/preview.mp4",
		PreviewVideoHLSURL: "https://example.com/preview.m3u8",
		PreviewImages:      pq.StringArray{"https://example.com/p1.jpg"},
		Maker:              "Existing Maker",
		Label:              "Existing Label",
		Series:             "Existing Series",
		Genres:             pq.StringArray{"genre1", "genre2"},
		Score:              4.5,
		Runtime:            120,
		ReleaseDate:        datatypes.Date(time.Date(2024, 2, 13, 0, 0, 0, 0, time.UTC)),
	}
}

func TestMovieInfo_PreserveFrom_EmptyFreshKeepsAllExisting(t *testing.T) {
	t.Parallel()
	// Simulates the bug case: a fresh fetch returned all-empty optional
	// fields but the required (IsValid) fields are populated. Without
	// PreserveFrom, this would clobber the DB. With PreserveFrom, all the
	// optional fields are restored from `existing`.
	fresh := &MovieInfo{
		ID:       "id-1",
		Number:   "NUM-001",
		Title:    "Existing Title",
		CoverURL: "https://example.com/cover.jpg",
		Provider: "TestProvider",
		Homepage: "https://example.com/id-1",
	}
	existing := fullExisting()

	fresh.PreserveFrom(existing)

	if fresh.Summary != existing.Summary {
		t.Errorf("Summary should be preserved from existing, got %q want %q", fresh.Summary, existing.Summary)
	}
	if fresh.Director != existing.Director {
		t.Errorf("Director should be preserved, got %q", fresh.Director)
	}
	if fresh.ThumbURL != existing.ThumbURL {
		t.Errorf("ThumbURL should be preserved, got %q", fresh.ThumbURL)
	}
	if fresh.BigThumbURL != existing.BigThumbURL {
		t.Errorf("BigThumbURL should be preserved, got %q", fresh.BigThumbURL)
	}
	if fresh.BigCoverURL != existing.BigCoverURL {
		t.Errorf("BigCoverURL should be preserved, got %q", fresh.BigCoverURL)
	}
	if fresh.PreviewVideoURL != existing.PreviewVideoURL {
		t.Errorf("PreviewVideoURL should be preserved, got %q", fresh.PreviewVideoURL)
	}
	if fresh.PreviewVideoHLSURL != existing.PreviewVideoHLSURL {
		t.Errorf("PreviewVideoHLSURL should be preserved, got %q", fresh.PreviewVideoHLSURL)
	}
	if fresh.Maker != existing.Maker {
		t.Errorf("Maker should be preserved, got %q", fresh.Maker)
	}
	if fresh.Label != existing.Label {
		t.Errorf("Label should be preserved, got %q", fresh.Label)
	}
	if fresh.Series != existing.Series {
		t.Errorf("Series should be preserved, got %q", fresh.Series)
	}
	if len(fresh.Actors) != len(existing.Actors) {
		t.Errorf("Actors should be preserved, got %v", fresh.Actors)
	}
	if len(fresh.PreviewImages) != len(existing.PreviewImages) {
		t.Errorf("PreviewImages should be preserved, got %v", fresh.PreviewImages)
	}
	if len(fresh.Genres) != len(existing.Genres) {
		t.Errorf("Genres should be preserved, got %v", fresh.Genres)
	}
	if fresh.Score != existing.Score {
		t.Errorf("Score should be preserved, got %v", fresh.Score)
	}
	if fresh.Runtime != existing.Runtime {
		t.Errorf("Runtime should be preserved, got %v", fresh.Runtime)
	}
	if time.Time(fresh.ReleaseDate) != time.Time(existing.ReleaseDate) {
		t.Errorf("ReleaseDate should be preserved, got %v", time.Time(fresh.ReleaseDate))
	}
}

func TestMovieInfo_PreserveFrom_NonEmptyFreshWinsOverExisting(t *testing.T) {
	t.Parallel()
	// "Non-empty wins" semantics: if the fresh fetch has a value, it must
	// override the existing DB value (provider has authority over its own data).
	fresh := &MovieInfo{
		ID:                 "id-1",
		Number:             "NUM-001",
		Title:              "Fresh Title",
		Summary:            "Fresh summary, brand new.",
		CoverURL:           "https://fresh.example.com/cover.jpg",
		Provider:           "TestProvider",
		Homepage:           "https://example.com/id-1",
		Director:           "Fresh Director",
		Actors:             pq.StringArray{"fresh-actor"},
		ThumbURL:           "https://fresh.example.com/thumb.jpg",
		BigThumbURL:        "https://fresh.example.com/big-thumb.jpg",
		BigCoverURL:        "https://fresh.example.com/big-cover.jpg",
		PreviewVideoURL:    "https://fresh.example.com/preview.mp4",
		PreviewVideoHLSURL: "https://fresh.example.com/preview.m3u8",
		PreviewImages:      pq.StringArray{"https://fresh.example.com/p1.jpg"},
		Maker:              "Fresh Maker",
		Label:              "Fresh Label",
		Series:             "Fresh Series",
		Genres:             pq.StringArray{"fresh-genre"},
		Score:              3.0,
		Runtime:            90,
		ReleaseDate:        datatypes.Date(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)),
	}
	existing := fullExisting()
	want := *fresh // snapshot before PreserveFrom

	fresh.PreserveFrom(existing)

	if fresh.Summary != want.Summary {
		t.Errorf("Summary: fresh should win, got %q want %q", fresh.Summary, want.Summary)
	}
	if fresh.Director != want.Director {
		t.Errorf("Director: fresh should win, got %q want %q", fresh.Director, want.Director)
	}
	if fresh.ThumbURL != want.ThumbURL {
		t.Errorf("ThumbURL: fresh should win, got %q", fresh.ThumbURL)
	}
	if fresh.Maker != want.Maker {
		t.Errorf("Maker: fresh should win, got %q", fresh.Maker)
	}
	if len(fresh.Actors) != 1 || fresh.Actors[0] != "fresh-actor" {
		t.Errorf("Actors: fresh should win (no merge/append), got %v", fresh.Actors)
	}
	if fresh.Score != 3.0 {
		t.Errorf("Score: fresh should win, got %v", fresh.Score)
	}
	if fresh.Runtime != 90 {
		t.Errorf("Runtime: fresh should win, got %v", fresh.Runtime)
	}
	if time.Time(fresh.ReleaseDate) != time.Time(want.ReleaseDate) {
		t.Errorf("ReleaseDate: fresh should win, got %v", time.Time(fresh.ReleaseDate))
	}
}

func TestMovieInfo_PreserveFrom_MixedFields(t *testing.T) {
	t.Parallel()
	// Real-world case: fresh has *some* fields (e.g. it got Title and Summary
	// but missed Director and Genres). Only the empty ones should be filled
	// from existing.
	fresh := &MovieInfo{
		ID:       "id-1",
		Number:   "NUM-001",
		Title:    "Fresh Title",
		Summary:  "Fresh summary kept.",
		CoverURL: "https://fresh.example.com/cover.jpg",
		Provider: "TestProvider",
		Homepage: "https://example.com/id-1",
		// Director, Genres, Score, Runtime, ReleaseDate intentionally empty.
		Actors: pq.StringArray{"fresh-actor"},
	}
	existing := fullExisting()

	fresh.PreserveFrom(existing)

	// Fresh-set fields should remain unchanged.
	if fresh.Summary != "Fresh summary kept." {
		t.Errorf("Summary was unexpectedly overwritten, got %q", fresh.Summary)
	}
	if len(fresh.Actors) != 1 || fresh.Actors[0] != "fresh-actor" {
		t.Errorf("Actors was unexpectedly overwritten, got %v", fresh.Actors)
	}

	// Empty fields should be filled from existing.
	if fresh.Director != existing.Director {
		t.Errorf("Director should be preserved, got %q", fresh.Director)
	}
	if len(fresh.Genres) != len(existing.Genres) {
		t.Errorf("Genres should be preserved, got %v", fresh.Genres)
	}
	if fresh.Score != existing.Score {
		t.Errorf("Score should be preserved, got %v", fresh.Score)
	}
	if fresh.Runtime != existing.Runtime {
		t.Errorf("Runtime should be preserved, got %v", fresh.Runtime)
	}
	if time.Time(fresh.ReleaseDate) != time.Time(existing.ReleaseDate) {
		t.Errorf("ReleaseDate should be preserved, got %v", time.Time(fresh.ReleaseDate))
	}
}

func TestMovieInfo_PreserveFrom_RequiredFieldsNeverTouched(t *testing.T) {
	t.Parallel()
	// Even if existing differs on a required (IsValid) field, PreserveFrom
	// must leave m's required fields alone. The fresh fetch is authoritative
	// for ID/Number/Title/CoverURL/Provider/Homepage.
	fresh := &MovieInfo{
		ID:       "fresh-id",
		Number:   "FRESH-001",
		Title:    "Fresh Title",
		CoverURL: "https://fresh.example.com/cover.jpg",
		Provider: "FreshProvider",
		Homepage: "https://fresh.example.com/id",
	}
	existing := &MovieInfo{
		ID:       "stale-id",
		Number:   "STALE-001",
		Title:    "Stale Title",
		CoverURL: "https://stale.example.com/cover.jpg",
		Provider: "StaleProvider",
		Homepage: "https://stale.example.com/id",
		Summary:  "stale summary",
	}

	fresh.PreserveFrom(existing)

	if fresh.ID != "fresh-id" {
		t.Errorf("ID should NEVER be replaced, got %q", fresh.ID)
	}
	if fresh.Number != "FRESH-001" {
		t.Errorf("Number should NEVER be replaced, got %q", fresh.Number)
	}
	if fresh.Title != "Fresh Title" {
		t.Errorf("Title should NEVER be replaced, got %q", fresh.Title)
	}
	if fresh.CoverURL != "https://fresh.example.com/cover.jpg" {
		t.Errorf("CoverURL should NEVER be replaced, got %q", fresh.CoverURL)
	}
	if fresh.Provider != "FreshProvider" {
		t.Errorf("Provider should NEVER be replaced, got %q", fresh.Provider)
	}
	if fresh.Homepage != "https://fresh.example.com/id" {
		t.Errorf("Homepage should NEVER be replaced, got %q", fresh.Homepage)
	}
	// Optional field: should still flow from existing.
	if fresh.Summary != "stale summary" {
		t.Errorf("Summary (optional) should be preserved from existing, got %q", fresh.Summary)
	}
}

func TestMovieInfo_PreserveFrom_EmptyArrayPreservesExisting(t *testing.T) {
	t.Parallel()
	// Edge case: nil vs empty slice. Both must be treated as "no value" so
	// the existing array is kept.
	for _, tc := range []struct {
		name   string
		actors pq.StringArray
	}{
		{"nil slice", nil},
		{"empty slice", pq.StringArray{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fresh := &MovieInfo{
				ID: "id-1", Number: "NUM-001", Title: "T",
				CoverURL: "c", Provider: "P", Homepage: "h",
				Actors: tc.actors,
			}
			existing := fullExisting()
			fresh.PreserveFrom(existing)
			if len(fresh.Actors) != len(existing.Actors) {
				t.Errorf("Actors should be preserved when fresh is %s, got %v", tc.name, fresh.Actors)
			}
		})
	}
}
