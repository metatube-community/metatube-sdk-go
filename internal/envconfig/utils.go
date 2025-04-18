package envconfig

import (
	"strings"
)

func normalizeProviderKey(key string) string {
	return strings.ToUpper(
		strings.ReplaceAll(key, "-", "_"))
}

func normalizeConfigKey(key string) string {
	return strings.ToUpper(key)
}
