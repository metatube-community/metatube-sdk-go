package babepedia

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestBabepediaActor_SearchActor(t *testing.T) {
	testkit.Test(t, New, []string{
		"Freya",
	})
}

func TestBabepediaActor_GetActorInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"Freya Parker",
	})
}
