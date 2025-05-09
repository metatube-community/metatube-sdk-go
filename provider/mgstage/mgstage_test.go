package mgstage

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestMGS_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"300MAAN-778",
	},
		testkit.FieldsNotEmpty("preview_images", "actors"),
		testkit.FieldsNotEmptyAny("maker", "label", "series"),
		testkit.FieldsNotEmptyAny("preview_video_url", "preview_video_hls_url"),
	)
}

func TestMGS_SearchMovie(t *testing.T) {
	testkit.Test(t, New, []string{
		"ABP-177",
		"200GANA-2701",
		"ABF-228",
	})
}

func TestMGS_GetMovieReviewsByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"200GANA-2701",
		"MAAN-930",
		"300MIUM-973",
		"MAAN-929",
		"ABP-177",
	})
}
