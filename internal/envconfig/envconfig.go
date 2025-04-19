package envconfig

import (
	"fmt"
	"maps"
	"os"
	"strings"
)

const (
	metaTubeEnvPrefix = "MT_"
	metaTubeConfigSep = "__"
)

var (
	metaTubeEnvs map[string]string

	ActorProviderConfigs ProviderConfigs
	MovieProviderConfigs ProviderConfigs
)

func init() {
	InitAllConfigs()
}

func InitAllConfigs() {
	initMetaTubeEnvs()
	ActorProviderConfigs = initProviderConfigs("actor")
	MovieProviderConfigs = initProviderConfigs("movie")
}

func initMetaTubeEnvs() {
	// reset before loading.
	metaTubeEnvs = make(map[string]string)
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
			metaTubeEnvs[key] = value
		}
	}
}

func initProviderConfigs(providerType string) ProviderConfigs {
	typed := parseProviderEnvsWithPrefix(
		fmt.Sprintf("%s%s_PROVIDER_", metaTubeEnvPrefix, strings.ToUpper(providerType)))
	common := parseProviderEnvsWithPrefix(
		fmt.Sprintf("%sPROVIDER_", metaTubeEnvPrefix)) // no type prefix
	return mergeProviderConfigs(typed, common)
}

func parseProviderEnvsWithPrefix(prefix string) providerMap {
	result := make(providerMap)
	for key, value := range metaTubeEnvs {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		trimmed := strings.TrimPrefix(key, prefix)
		provider, configKey, found := strings.Cut(trimmed, metaTubeConfigSep)
		if !found {
			continue // malformed entry.
		}
		provider = normalizeProviderKey(provider)
		configKey = normalizeConfigKey(configKey)

		if _, ok := result[provider]; !ok {
			result[provider] = make(providerSetting)
		}
		result[provider][configKey] = value
	}
	return result
}

func mergeProviderConfigs(primary, fallback providerMap) ProviderConfigs {
	merged := make(providerMap)
	for provider, setting := range fallback {
		merged[provider] = maps.Clone(setting)
	}
	for provider, setting := range primary {
		if _, exists := merged[provider]; !exists {
			merged[provider] = make(providerSetting)
		}
		for k, v := range setting {
			// primarily map overwrites fallback.
			merged[provider][k] = v
		}
	}
	return ProviderConfigs{merged}
}
