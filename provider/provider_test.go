package provider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider_GetMovieInfo(t *testing.T) {
	for _, unit := range []struct {
		builder func() Provider
		movieID string
	}{
		{NewJavBus, "ABP-331"},
	} {
		provider := unit.builder()
		info, err := provider.GetMovieInfo(unit.movieID)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.Nil(t, err)
		t.Logf("%s", data)
	}
}

func TestProvider_SearchMovie(t *testing.T) {
	for _, unit := range []struct {
		builder func() Provider
		keyword string
	}{
		{NewJavBus, "abp-331"},
	} {
		provider := unit.builder()
		result, err := provider.SearchMovie(unit.keyword)
		data, _ := json.MarshalIndent(result, "", "\t")
		assert.Nil(t, err)
		t.Logf("%s", data)
	}
}
