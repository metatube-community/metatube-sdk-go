package theporndb

import (
	"os"
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

// Set env MT_THEPORNDB_ACCESS_TOKEN to run tests.
var accessToken = os.Getenv("MT_THEPORNDB_ACCESS_TOKEN")

func TestThePornDBVideo_GetMovieInfoByID(t *testing.T) {
	if accessToken == "" {
		t.Skip("MT_THEPORNDB_ACCESS_TOKEN is not set")
	}
	testkit.Test(t, func() *ThePornDBVideo {
		res := NewThePornDBScene()
		res.accessToken = accessToken
		return res
	}, []string{
		"i-want-clips-leaking-into-debt",
		"f32b7a83-9477-4c6f-8d8a-e60dae1aada3",
		"6474100",
	})

	testkit.Test(t, func() *ThePornDBVideo {
		res := NewThePornDBMovie()
		res.accessToken = accessToken
		return res
	}, []string{
		"digital-sin-sisterly-love-4",
		"ee8cf0e2-55b3-41fa-a7f6-7db83068b2e4",
		"6472937",
	})
}

func TestThePornDBVideo_GetMovieInfoByURL(t *testing.T) {
	if accessToken == "" {
		t.Skip("MT_THEPORNDB_ACCESS_TOKEN is not set")
	}
	testkit.Test(t, func() *ThePornDBVideo {
		res := NewThePornDBScene()
		res.accessToken = accessToken
		return res
	}, []string{
		sceneBaseURL + "i-want-clips-leaking-into-debt",
		sceneBaseURL + "f32b7a83-9477-4c6f-8d8a-e60dae1aada3",
		sceneBaseURL + "6474100",
	})

	testkit.Test(t, func() *ThePornDBVideo {
		res := NewThePornDBMovie()
		res.accessToken = accessToken
		return res
	}, []string{
		movieBaseURL + "digital-sin-sisterly-love-4",
		movieBaseURL + "ee8cf0e2-55b3-41fa-a7f6-7db83068b2e4",
		movieBaseURL + "6472937",
	})
}

func TestThePornDBVideo_SearchMovie(t *testing.T) {
	if accessToken == "" {
		t.Skip("MT_THEPORNDB_ACCESS_TOKEN is not set")
	}
	testkit.Test(t, func() *ThePornDBVideo {
		res := NewThePornDBScene()
		res.accessToken = accessToken
		return res
	}, []string{
		"The Three Evil Dragon",
		"6377406",
		// search on slug and uuid does not work.
	})

	testkit.Test(t, func() *ThePornDBVideo {
		res := NewThePornDBMovie()
		res.accessToken = accessToken
		return res
	}, []string{
		"Sisterly Love 4",
		"6472937",
		// search on slug and uuid does not work.
	})
}

func TestThePornDBActor_GetActorInfoByID(t *testing.T) {
	if accessToken == "" {
		t.Skip("MT_THEPORNDB_ACCESS_TOKEN is not set")
	}
	testkit.Test(t, func() *ThePornDBActor {
		res := NewThePornDBActor()
		res.accessToken = accessToken
		return res
	}, []string{
		"adf8435e-d5df-42b9-b46b-8440dee5a271",
		"harley-king",
		"138309",
	})
}

func TestThePornDBActor_GetActorInfoByURL(t *testing.T) {
	if accessToken == "" {
		t.Skip("MT_THEPORNDB_ACCESS_TOKEN is not set")
	}
	testkit.Test(t, func() *ThePornDBActor {
		res := NewThePornDBActor()
		res.accessToken = accessToken
		return res
	}, []string{
		actorBaseURL + "adf8435e-d5df-42b9-b46b-8440dee5a271",
		actorBaseURL + "harley-king",
		actorBaseURL + "138309",
	})
}

func TestThePornDBActor_SearchActor(t *testing.T) {
	if accessToken == "" {
		t.Skip("MT_THEPORNDB_ACCESS_TOKEN is not set")
	}
	testkit.Test(t, func() *ThePornDBActor {
		res := NewThePornDBActor()
		res.accessToken = accessToken
		return res
	}, []string{
		"Harley",
		"138309",
		"harley-king",
		"adf8435e-d5df-42b9-b46b-8440dee5a271",
	})
}
