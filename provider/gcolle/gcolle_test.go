package gcolle

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestGcolle_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"847256",
		"848234",
		"845371",
		"839979",
		"848315",
	})
}
