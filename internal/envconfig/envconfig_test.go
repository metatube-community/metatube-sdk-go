package envconfig

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaTubeEnvConfigs(t *testing.T) {
	// Set mock environment variables
	os.Clearenv()
	for _, unit := range []struct {
		key, value string
	}{
		{"MT_ACTOR_PROVIDER_ABC__PRIORITY", "1"},
		{"MT_MOVIE_PROVIDER_DEF__PRIORITY", "2"},
		{"MT_MOVIE_PROVIDER_xyz__PRIORITY", "0"},
		{"MT_MOVIE_PROVIDER_OPQ__TOKEN", "token1"},
		{"MT_MOVIE_PROVIDER_UVW__tOkEn", "token2"},
		{"MT_PROVIDER_UVW__TIMEout", "900s"},
		{"MT_PROVIDER_UVW__PRIORITY", "5"},       // -> actor/movie
		{"MT_MOVIE_PROVIDER_UVW__PRIORITY", "0"}, // override movie
		{"irrelevant_key", "ignore_me"},
		{"mt_malformed_key", "ignore_me"},
	} {
		err := os.Setenv(unit.key, unit.value)
		require.NoError(t, err)
	}

	// Reload
	InitAllConfigs()

	// Assert
	assert.Len(t, metaTubeEnvs, 8+1 /* includes malformed */)
	assert.Equal(t, ActorProviderConfigs.IsEmpty(), false)
	assert.Equal(t, MovieProviderConfigs.IsEmpty(), false)

	for k, v := range map[string]float64{
		"ABC": 1,
		"UVW": 5,
	} {
		priority, ok := ActorProviderConfigs.GetPriority(k)
		if assert.True(t, ok) {
			assert.Equal(t, v, priority)
		}
	}

	for k, v := range map[string]float64{
		"DEF": 2,
		"XYZ": 0,
		"UVW": 0,
	} {
		priority, ok := MovieProviderConfigs.GetPriority(k)
		if assert.True(t, ok) {
			assert.Equal(t, v, priority)
		}
	}

	val, _ := MovieProviderConfigs.Get("OPQ", "token")
	assert.Equal(t, "token1", val)

	val, _ = MovieProviderConfigs.Get("UVW", "ToKeN")
	assert.Equal(t, "token2", val)

	timeout, _ := MovieProviderConfigs.GetTimeout("UVW")
	assert.Equal(t, 900*time.Second, timeout)
}
