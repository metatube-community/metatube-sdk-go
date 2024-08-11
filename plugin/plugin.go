package plugin

import (
	"sync"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/model"
)

type (
	ActorInfoPlugin func(*engine.Engine, *model.ActorInfo)
	MovieInfoPlugin func(*engine.Engine, *model.MovieInfo)
)

var (
	actorPluginsMu sync.RWMutex
	moviePluginsMu sync.RWMutex

	// ActorInfoPlugin & MovieInfoPlugin Registry
	actorPlugins = make(map[string]ActorInfoPlugin)
	moviePlugins = make(map[string]MovieInfoPlugin)
)

func RegisterActorInfoPlugin(name string, plugin ActorInfoPlugin) {
	actorPluginsMu.Lock()
	actorPlugins[name] = plugin
	actorPluginsMu.Unlock()
}

func LookupActorInfoPlugin(name string) (ActorInfoPlugin, bool) {
	actorPluginsMu.RLock()
	defer actorPluginsMu.RUnlock()
	plugin, ok := actorPlugins[name]
	return plugin, ok
}

func RegisterMovieInfoPlugin(name string, plugin MovieInfoPlugin) {
	moviePluginsMu.Lock()
	moviePlugins[name] = plugin
	moviePluginsMu.Unlock()
}

func LookupMovieInfoPlugin(name string) (MovieInfoPlugin, bool) {
	moviePluginsMu.RLock()
	defer moviePluginsMu.RUnlock()
	plugin, ok := moviePlugins[name]
	return plugin, ok
}
