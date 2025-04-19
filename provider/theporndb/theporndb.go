package theporndb

import (
	"os"

	"github.com/metatube-community/metatube-sdk-go/provider"
)

// ThePornDB is disabled by default, use `export MT_THEPORNDB_ACCESS_TOKEN=your-token` to enable.
var accessToken = os.Getenv("MT_THEPORNDB_ACCESS_TOKEN")

const Priority = 1000

func init() {
	provider.Register(SceneProviderName, NewThePornDBScene)
	provider.Register(MovieProviderName, NewThePornDBMovie)
	provider.Register(ActorProviderName, NewThePornDBActor)
}
