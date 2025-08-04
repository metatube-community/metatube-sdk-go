package theporndb

import (
	"github.com/metatube-community/metatube-sdk-go/provider"
)

// ThePornDB is disabled by default, to enable:
// `export MT_ACTOR_PROVIDER_THEPORNDBACTOR__ACCESS_TOKEN=your-token`
// `export MT_MOVIE_PROVIDER_THEPORNDBMOVIE__ACCESS_TOKEN=your-token`
// `export MT_MOVIE_PROVIDER_THEPORNDBSCENE__ACCESS_TOKEN=your-token`

const Priority = 1000

func init() {
	provider.Register(SceneProviderName, NewThePornDBScene)
	provider.Register(MovieProviderName, NewThePornDBMovie)
	provider.Register(ActorProviderName, NewThePornDBActor)
}
