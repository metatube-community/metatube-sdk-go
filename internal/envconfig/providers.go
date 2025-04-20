package envconfig

import (
	"maps"
	"strconv"
	"time"

	"github.com/metatube-community/metatube-sdk-go/provider"
)

var _ provider.ConfigGetter = (*ProviderConfigs)(nil)

type (
	providerSetting = map[string]string
	providerMap     = map[string]providerSetting
)

type ProviderConfigs struct {
	providers providerMap
}

func (c ProviderConfigs) IsEmpty() bool {
	return len(c.providers) == 0
}

func (c ProviderConfigs) Clone() ProviderConfigs {
	return ProviderConfigs{maps.Clone(c.providers)}
}

func (c ProviderConfigs) Get(provider, key string) (string, bool) {
	if config, ok := c.providers[normalizeProviderKey(provider)]; ok {
		val, found := config[normalizeConfigKey(key)]
		return val, found
	}
	return "", false
}

func (c ProviderConfigs) GetConfig(provider string) map[string]string {
	return maps.Clone(
		c.providers[normalizeProviderKey(provider)],
	)
}

func (c ProviderConfigs) GetPriority(provider string) (float64, bool) {
	value, ok := c.Get(provider, "PRIORITY")
	priority, _ := strconv.ParseFloat(value, 64)
	return priority, ok
}

func (c ProviderConfigs) GetTimeout(provider string) (time.Duration, bool) {
	value, ok := c.Get(provider, "TIMEOUT")
	if !ok {
		return 0, false
	}
	priority, err := time.ParseDuration(value)
	return priority, err == nil
}
