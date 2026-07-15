package model

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

func TestActorInfo_IsValid(t *testing.T) {
	t.Parallel()

	base := ActorInfo{
		ID:       "a1",
		Name:     "Alice",
		Provider: "Test",
		Homepage: "https://example.com/a1",
	}
	if !base.IsValid() {
		t.Fatalf("baseline ActorInfo should be valid, got invalid: %+v", base)
	}

	withoutSummary := base
	withoutSummary.Summary = ""
	if !withoutSummary.IsValid() {
		t.Errorf("ActorInfo with empty Summary should still be valid (Summary is optional)")
	}

	for _, tc := range []struct {
		name string
		mut  func(*ActorInfo)
	}{
		{"missing ID", func(a *ActorInfo) { a.ID = "" }},
		{"missing Name", func(a *ActorInfo) { a.Name = "" }},
		{"missing Provider", func(a *ActorInfo) { a.Provider = "" }},
		{"missing Homepage", func(a *ActorInfo) { a.Homepage = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := base
			tc.mut(&a)
			if a.IsValid() {
				t.Errorf("ActorInfo missing required field should be invalid: %+v", a)
			}
		})
	}
}

func TestActorInfo_PreserveFrom_NilExisting(t *testing.T) {
	t.Parallel()
	a := &ActorInfo{ID: "a1"}
	a.PreserveFrom(nil)
	if a.ID != "a1" {
		t.Errorf("PreserveFrom(nil) should not touch any field, got ID=%q", a.ID)
	}
}

func fullExistingActor() *ActorInfo {
	return &ActorInfo{
		ID:           "a1",
		Name:         "Existing Alice",
		Provider:     "TestProvider",
		Homepage:     "https://example.com/a1",
		Summary:      "Existing bio.",
		Hobby:        "Existing hobby",
		Skill:        "Existing skill",
		BloodType:    "A",
		CupSize:      "C",
		Measurements: "B85-W58-H86",
		Nationality:  "Japan",
		Height:       165,
		Aliases:      pq.StringArray{"alias1", "alias2"},
		Images:       pq.StringArray{"https://example.com/img1.jpg"},
		Birthday:     datatypes.Date(time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC)),
		DebutDate:    datatypes.Date(time.Date(2015, 6, 1, 0, 0, 0, 0, time.UTC)),
	}
}

func TestActorInfo_PreserveFrom_EmptyFreshKeepsAllExisting(t *testing.T) {
	t.Parallel()
	fresh := &ActorInfo{
		ID:       "a1",
		Name:     "Existing Alice",
		Provider: "TestProvider",
		Homepage: "https://example.com/a1",
	}
	existing := fullExistingActor()

	fresh.PreserveFrom(existing)

	if fresh.Summary != existing.Summary {
		t.Errorf("Summary should be preserved, got %q", fresh.Summary)
	}
	if fresh.Hobby != existing.Hobby {
		t.Errorf("Hobby should be preserved, got %q", fresh.Hobby)
	}
	if fresh.Skill != existing.Skill {
		t.Errorf("Skill should be preserved, got %q", fresh.Skill)
	}
	if fresh.BloodType != existing.BloodType {
		t.Errorf("BloodType should be preserved, got %q", fresh.BloodType)
	}
	if fresh.CupSize != existing.CupSize {
		t.Errorf("CupSize should be preserved, got %q", fresh.CupSize)
	}
	if fresh.Measurements != existing.Measurements {
		t.Errorf("Measurements should be preserved, got %q", fresh.Measurements)
	}
	if fresh.Nationality != existing.Nationality {
		t.Errorf("Nationality should be preserved, got %q", fresh.Nationality)
	}
	if fresh.Height != existing.Height {
		t.Errorf("Height should be preserved, got %d", fresh.Height)
	}
	if len(fresh.Aliases) != len(existing.Aliases) {
		t.Errorf("Aliases should be preserved, got %v", fresh.Aliases)
	}
	if len(fresh.Images) != len(existing.Images) {
		t.Errorf("Images should be preserved, got %v", fresh.Images)
	}
	if time.Time(fresh.Birthday) != time.Time(existing.Birthday) {
		t.Errorf("Birthday should be preserved, got %v", time.Time(fresh.Birthday))
	}
	if time.Time(fresh.DebutDate) != time.Time(existing.DebutDate) {
		t.Errorf("DebutDate should be preserved, got %v", time.Time(fresh.DebutDate))
	}
}

func TestActorInfo_PreserveFrom_NonEmptyFreshWinsOverExisting(t *testing.T) {
	t.Parallel()
	fresh := &ActorInfo{
		ID:           "a1",
		Name:         "Fresh Alice",
		Provider:     "TestProvider",
		Homepage:     "https://example.com/a1",
		Summary:      "Fresh bio.",
		Hobby:        "Fresh hobby",
		Skill:        "Fresh skill",
		BloodType:    "O",
		CupSize:      "D",
		Measurements: "B90-W60-H88",
		Nationality:  "Korea",
		Height:       170,
		Aliases:      pq.StringArray{"fresh-alias"},
		Images:       pq.StringArray{"https://fresh.example.com/img.jpg"},
		Birthday:     datatypes.Date(time.Date(1996, 2, 2, 0, 0, 0, 0, time.UTC)),
		DebutDate:    datatypes.Date(time.Date(2016, 7, 2, 0, 0, 0, 0, time.UTC)),
	}
	existing := fullExistingActor()
	want := *fresh

	fresh.PreserveFrom(existing)

	if fresh.Summary != want.Summary {
		t.Errorf("Summary: fresh should win, got %q", fresh.Summary)
	}
	if fresh.BloodType != want.BloodType {
		t.Errorf("BloodType: fresh should win, got %q", fresh.BloodType)
	}
	if fresh.Height != want.Height {
		t.Errorf("Height: fresh should win, got %d", fresh.Height)
	}
	if len(fresh.Aliases) != 1 || fresh.Aliases[0] != "fresh-alias" {
		t.Errorf("Aliases: fresh should win (no merge), got %v", fresh.Aliases)
	}
	if time.Time(fresh.Birthday) != time.Time(want.Birthday) {
		t.Errorf("Birthday: fresh should win, got %v", time.Time(fresh.Birthday))
	}
}

func TestActorInfo_PreserveFrom_RequiredFieldsNeverTouched(t *testing.T) {
	t.Parallel()
	fresh := &ActorInfo{
		ID:       "fresh-id",
		Name:     "Fresh Name",
		Provider: "FreshProvider",
		Homepage: "https://fresh.example.com/id",
	}
	existing := &ActorInfo{
		ID:       "stale-id",
		Name:     "Stale Name",
		Provider: "StaleProvider",
		Homepage: "https://stale.example.com/id",
		Summary:  "stale bio",
	}

	fresh.PreserveFrom(existing)

	if fresh.ID != "fresh-id" {
		t.Errorf("ID should NEVER be replaced, got %q", fresh.ID)
	}
	if fresh.Name != "Fresh Name" {
		t.Errorf("Name should NEVER be replaced, got %q", fresh.Name)
	}
	if fresh.Provider != "FreshProvider" {
		t.Errorf("Provider should NEVER be replaced, got %q", fresh.Provider)
	}
	if fresh.Homepage != "https://fresh.example.com/id" {
		t.Errorf("Homepage should NEVER be replaced, got %q", fresh.Homepage)
	}
	if fresh.Summary != "stale bio" {
		t.Errorf("Summary (optional) should be preserved, got %q", fresh.Summary)
	}
}
