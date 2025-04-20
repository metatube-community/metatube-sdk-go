package envconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/collection/maps"
)

const (
	metaTubeEnvPrefix = "MT_"
	metaTubeConfigSep = "__"
)

// metaTubeEnvs stores all MetaTube related environment variables.
var metaTubeEnvs *maps.CaseInsensitiveMap[string]

var (
	ActorProviderConfigs *maps.CaseInsensitiveMap[*Config]
	MovieProviderConfigs *maps.CaseInsensitiveMap[*Config]
)

func init() {
	InitAllEnvConfigs()
}

func InitAllEnvConfigs() {
	metaTubeEnvs = initMetaTubeEnvs()
	ActorProviderConfigs = initProviderConfigs("actor")
	MovieProviderConfigs = initProviderConfigs("movie")
}

func initMetaTubeEnvs() *maps.CaseInsensitiveMap[string] {
	envs := maps.NewCaseInsensitiveMap[string]()
	// initialize ENV.
	for _, env := range os.Environ() {
		key, value, found := strings.Cut(env, "=")
		if !found {
			continue
		} else {
			// always uppercase keys.
			key = strings.ToUpper(key)
		}
		if strings.HasPrefix(key, metaTubeEnvPrefix) {
			envs.Set(key, value)
		}
	}
	return envs
}

func initProviderConfigs(providerType string) *maps.CaseInsensitiveMap[*Config] {
	typed := parseProviderEnvsWithPrefix(
		fmt.Sprintf("%s%s_PROVIDER_", metaTubeEnvPrefix, strings.ToUpper(providerType)))
	common := parseProviderEnvsWithPrefix(
		fmt.Sprintf("%sPROVIDER_", metaTubeEnvPrefix)) // no type prefix
	return mergeProviderConfigs(typed, common)
}

func parseProviderEnvsWithPrefix(prefix string) *maps.CaseInsensitiveMap[*Config] {
	result := maps.NewCaseInsensitiveMap[*Config]()
	for key, value := range metaTubeEnvs.Iterator() {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		trimmed := strings.TrimPrefix(key, prefix)
		provider, configKey, found := strings.Cut(trimmed, metaTubeConfigSep)
		if !found {
			continue // malformed entry.
		}

		if !result.Has(provider) {
			result.Set(provider, NewConfig())
		}
		result.GetOrDefault(provider).Set(configKey, value)
	}
	return result
}

func mergeProviderConfigs(primary, fallback *maps.CaseInsensitiveMap[*Config]) *maps.CaseInsensitiveMap[*Config] {
	merged := maps.NewCaseInsensitiveMap[*Config]()
	for provider, config := range fallback.Iterator() {
		merged.Set(provider, config.Copy())
	}
	for provider, config := range primary.Iterator() {
		if !merged.Has(provider) {
			merged.Set(provider, NewConfig())
		}
		for k, v := range config.Iterator() {
			// primarily map overwrites fallback.
			merged.GetOrDefault(provider).Set(k, v)
		}
	}
	return merged
}
