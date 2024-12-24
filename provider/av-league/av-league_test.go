package avleague

import (
	"testing"

	"github.com/metatube-community/metatube-sdk-go/provider/internal/testkit"
)

func TestAVLeague_GetActorInfoByID(t *testing.T) {
	testkit.Test(t, New, []string{
		"8301",
		"14005",
		"36672",
		"32759",
		"36736",
	})
}

func TestAVLeague_SearchActor(t *testing.T) {
	testkit.Test(t, New, []string{
		"白川ゆず",
		"美竹すず",
		"宇流木さら",
	})
}
