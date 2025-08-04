package envconfig

import (
	"errors"
	"strconv"
	"time"

	"github.com/metatube-community/metatube-sdk-go/collection/maps"
	"github.com/metatube-community/metatube-sdk-go/provider"
)

var _ provider.Config = (*Config)(nil)

var errKeyNotFound = errors.New("key not found")

type Config struct {
	*maps.CaseInsensitiveMap[string]
}

func NewConfig() *Config {
	return &Config{maps.NewCaseInsensitiveMap[string]()}
}

func (c Config) Copy() *Config {
	return &Config{c.CaseInsensitiveMap.Copy()}
}

func (c Config) GetString(key string) (string, error) {
	v, ok := c.Get(key)
	if !ok {
		return "", errKeyNotFound
	}
	return v, nil
}

func (c Config) GetBool(key string) (bool, error) {
	v, ok := c.Get(key)
	if !ok {
		return false, errKeyNotFound
	}
	return strconv.ParseBool(v)
}

func (c Config) GetInt64(key string) (int64, error) {
	v, ok := c.Get(key)
	if !ok {
		return 0, errKeyNotFound
	}
	return strconv.ParseInt(v, 10, 64)
}

func (c Config) GetFloat64(key string) (float64, error) {
	v, ok := c.Get(key)
	if !ok {
		return 0, errKeyNotFound
	}
	return strconv.ParseFloat(v, 64)
}

func (c Config) GetDuration(key string) (time.Duration, error) {
	v, ok := c.Get(key)
	if !ok {
		return 0, errKeyNotFound
	}
	return time.ParseDuration(v)
}
