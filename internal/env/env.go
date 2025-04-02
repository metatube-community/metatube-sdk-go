package env

import (
	"os"
	"strconv"
	"strings"
)

const MetaTubeEnvPrefix = "MT_"

// Special environment prefixes for setting provider priorities.
const (
	actorProviderPriorityEnvPrefix = MetaTubeEnvPrefix + "ACTOR_PROVIDER_PRIORITY_"
	movieProviderPriorityEnvPrefix = MetaTubeEnvPrefix + "MOVIE_PROVIDER_PRIORITY_"
)

var _metaTubeEnvs map[string]string

func init() {
	loadMetaTubeEnvs()
}

func loadMetaTubeEnvs() {
	// reset before loading.
	_metaTubeEnvs = make(map[string]string)
	// initialize ENV.
	for _, env := range os.Environ() {
		key, value, found := strings.Cut(env, "=")
		if !found {
			continue
		} else {
			// always uppercase keys.
			key = strings.ToUpper(key)
		}
		if strings.HasPrefix(key, MetaTubeEnvPrefix) {
			_metaTubeEnvs[key] = value
		}
	}
}

func getProviderPriorities(prefix string) map[string]float64 {
	priorities := make(map[string]float64)
	for key, value := range _metaTubeEnvs {
		if strings.HasPrefix(key, prefix) {
			provider := key[len(prefix):]
			// TODO: improve this provider weight settings.
			provider = strings.ReplaceAll(provider, "_", "-")
			priority, _ := strconv.ParseFloat(value, 64)
			priorities[provider] = priority
		}
	}
	return priorities
}

func ActorProviderPriorities() map[string]float64 {
	return getProviderPriorities(actorProviderPriorityEnvPrefix)
}

func MovieProviderPriorities() map[string]float64 {
	return getProviderPriorities(movieProviderPriorityEnvPrefix)
}
