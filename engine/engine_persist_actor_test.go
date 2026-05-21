package engine

import (
	"net/url"
	"testing"
	"time"

	"github.com/lib/pq"
	"golang.org/x/text/language"
	"gorm.io/datatypes"

	mt "github.com/metatube-community/metatube-sdk-go/provider"

	"github.com/metatube-community/metatube-sdk-go/model"
)

// fakeActorProvider implements mt.ActorProvider for tests. Language is set to
// English by default so the gfriends image-injection defer in
// getActorInfoWithCallback is skipped (it only triggers for Japanese providers
// and would require a registered gfriends provider).
type fakeActorProvider struct {
	name string
	url  *url.URL
	lang language.Tag
	next *model.ActorInfo
	err  error
}

func (f *fakeActorProvider) Name() string                              { return f.name }
func (f *fakeActorProvider) Priority() float64                         { return 1 }
func (f *fakeActorProvider) SetPriority(v float64)                     {}
func (f *fakeActorProvider) Language() language.Tag                    { return f.lang }
func (f *fakeActorProvider) URL() *url.URL                             { return f.url }
func (f *fakeActorProvider) NormalizeActorID(id string) string         { return id }
func (f *fakeActorProvider) ParseActorIDFromURL(string) (string, error) {
	return "", nil
}
func (f *fakeActorProvider) GetActorInfoByID(id string) (*model.ActorInfo, error) {
	return f.next, f.err
}
func (f *fakeActorProvider) GetActorInfoByURL(string) (*model.ActorInfo, error) {
	return f.next, f.err
}

var _ mt.ActorProvider = (*fakeActorProvider)(nil)

func newFakeActorProvider(name string) *fakeActorProvider {
	u, _ := url.Parse("https://example.com/")
	return &fakeActorProvider{name: name, url: u, lang: language.English}
}

func fullActorInfo(id, provider string) *model.ActorInfo {
	return &model.ActorInfo{
		ID:           id,
		Name:         "Existing Alice",
		Provider:     provider,
		Homepage:     "https://example.com/" + id,
		Summary:      "Full bio.",
		Hobby:        "Hobby X",
		Skill:        "Skill Y",
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

func minimalActorInfo(id, provider string) *model.ActorInfo {
	return &model.ActorInfo{
		ID:       id,
		Name:     "Existing Alice",
		Provider: provider,
		Homepage: "https://example.com/" + id,
	}
}

// TestGetActorInfo_EmptyFieldsDoNotClobberExisting is the actor-side analog of
// TestGetMovieInfo_EmptyFieldsDoNotClobberExisting. It verifies that a fetch
// returning only the IsValid() fields does not wipe the actor's Summary,
// Aliases, Images, Birthday, etc. from the DB.
func TestGetActorInfo_EmptyFieldsDoNotClobberExisting(t *testing.T) {
	e := newTestEngine(t)
	p := newFakeActorProvider("FakeActorProvider")

	id := "a-1"

	// --- Step 1: full insert.
	full := fullActorInfo(id, p.name)
	p.next = full
	if _, err := e.getActorInfoByProviderID(p, id, false /* lazy */); err != nil {
		t.Fatalf("step 1: %v", err)
	}

	stored, err := e.getActorInfoFromDB(p, id)
	if err != nil {
		t.Fatalf("DB read after step 1: %v", err)
	}
	if stored.Summary != full.Summary {
		t.Fatalf("step 1: DB Summary mismatch")
	}

	// --- Step 2: provider returns minimal info.
	p.next = minimalActorInfo(id, p.name)
	if _, err := e.getActorInfoByProviderID(p, id, false /* lazy */); err != nil {
		t.Fatalf("step 2: %v", err)
	}

	stored2, err := e.getActorInfoFromDB(p, id)
	if err != nil {
		t.Fatalf("DB read after step 2: %v", err)
	}
	if stored2.Summary != full.Summary {
		t.Errorf("step 2: DB Summary was clobbered, got %q want %q",
			stored2.Summary, full.Summary)
	}
	if stored2.Hobby != full.Hobby {
		t.Errorf("step 2: DB Hobby was clobbered, got %q", stored2.Hobby)
	}
	if stored2.BloodType != full.BloodType {
		t.Errorf("step 2: DB BloodType was clobbered, got %q", stored2.BloodType)
	}
	if stored2.Height != full.Height {
		t.Errorf("step 2: DB Height was clobbered, got %d", stored2.Height)
	}
	if len(stored2.Aliases) != len(full.Aliases) {
		t.Errorf("step 2: DB Aliases was clobbered, got %v", stored2.Aliases)
	}
	if len(stored2.Images) != len(full.Images) {
		t.Errorf("step 2: DB Images was clobbered, got %v", stored2.Images)
	}
	if time.Time(stored2.Birthday) != time.Time(full.Birthday) {
		t.Errorf("step 2: DB Birthday was clobbered, got %v", time.Time(stored2.Birthday))
	}
	if time.Time(stored2.DebutDate) != time.Time(full.DebutDate) {
		t.Errorf("step 2: DB DebutDate was clobbered, got %v", time.Time(stored2.DebutDate))
	}
}

// TestGetActorInfo_NonEmptyFieldsWinOverExisting confirms the converse for
// actors: a non-empty fresh fetch replaces the corresponding DB column.
func TestGetActorInfo_NonEmptyFieldsWinOverExisting(t *testing.T) {
	e := newTestEngine(t)
	p := newFakeActorProvider("FakeActorProvider")

	id := "a-1"

	p.next = fullActorInfo(id, p.name)
	if _, err := e.getActorInfoByProviderID(p, id, false); err != nil {
		t.Fatalf("seed: %v", err)
	}

	v2 := fullActorInfo(id, p.name)
	v2.Summary = "Updated bio - v2."
	v2.Height = 170
	v2.Aliases = pq.StringArray{"new-alias"}
	p.next = v2
	if _, err := e.getActorInfoByProviderID(p, id, false); err != nil {
		t.Fatalf("v2: %v", err)
	}

	stored, err := e.getActorInfoFromDB(p, id)
	if err != nil {
		t.Fatalf("DB read: %v", err)
	}
	if stored.Summary != "Updated bio - v2." {
		t.Errorf("Summary v2 should win, got %q", stored.Summary)
	}
	if stored.Height != 170 {
		t.Errorf("Height v2 should win, got %d", stored.Height)
	}
	if len(stored.Aliases) != 1 || stored.Aliases[0] != "new-alias" {
		t.Errorf("Aliases v2 should replace v1, got %v", stored.Aliases)
	}
}
