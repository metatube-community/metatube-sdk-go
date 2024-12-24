package c0930

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestC0930_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"ki220913",
		"hitozuma1391",
		"hitozuma1371",
	})
}
