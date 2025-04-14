package gfriends

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestGfriends_GetActorImagesByName(t *testing.T) {
	testkit.Test(t, New, []string{
		"美竹すず",
	})
}
