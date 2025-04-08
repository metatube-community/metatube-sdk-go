package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaTubeEnvs(t *testing.T) {
	// Set mock environment variables
	os.Clearenv()
	os.Setenv("MT_ACTOR_PROVIDER_PRIORITY_ABC", "1")
	os.Setenv("MT_MOVIE_PROVIDER_PRIORITY_DEF", "2")
	os.Setenv("MT_MOVIE_PROVIDER_PRIORITY_xyz", "0")
	os.Setenv("irrelevant_key", "ignore_me")

	// Reload
	loadMetaTubeEnvs()

	// Assert
	got := _metaTubeEnvs
	assert.Len(t, got, 3)

	ap := ActorProviderPriorities()
	assert.Len(t, ap, 1)
	assert.Equal(t, ap, map[string]float64{
		"ABC": 1,
	})

	mp := MovieProviderPriorities()
	assert.Len(t, mp, 2)
	assert.Equal(t, mp, map[string]float64{
		"DEF": 2,
		"XYZ": 0,
	})
}
