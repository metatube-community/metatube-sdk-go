package aylo

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestBrazzers_SearchMovie(t *testing.T) {
	testkit.Test(t, NewBrazzers, []string{
		"cumming.back.for.more",
		"BrazzersExxtra.25.09.19.Amirah.Adara.Door.Cam.Catches.Dildo.Cheater.XXX.1080p.HEVC.x265.PRT.mkv",
	})
}
