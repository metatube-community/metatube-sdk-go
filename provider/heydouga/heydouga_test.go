package heydouga

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestHeyDouga_GetMovieInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		// The IDs below are real ones observed on the site's homepage, covering
		// three slug shapes:
		//   plain number   {brand}-{number}        e.g. 4382-450
		//   hyphen         {brand}-ppv-date-number  e.g. 3001-ppv-010325-001
		//   underscore     {brand}-ppv-date_number  e.g. 3002-ppv-010324_001
		//   alpha-mixed    {brand}-r019
		"4382-450",
		"4080-2489",
		"3001-ppv-010325-001",
		"3002-ppv-010324_001",
		"4198-r019",
	})
}
