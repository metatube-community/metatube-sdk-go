package provider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvider_GetMovieInfoByID(t *testing.T) {
	for _, unit := range []struct {
		builder func() Provider
		movieID string
	}{
		//{NewJavBus, "ABP-331"},
		//{NewMGStage, "277DCV-197"},
		{NewDMM, "h_198need00094r18"},
	} {
		provider := unit.builder()
		info, err := provider.GetMovieInfoByID(unit.movieID)
		data, _ := json.MarshalIndent(info, "", "\t")
		assert.Nil(t, err)
		assert.True(t, info.Valid())
		t.Logf("%s", data)
	}
}

func TestProvider_SearchMovie(t *testing.T) {
	for _, unit := range []struct {
		builder func() Provider
		keyword string
	}{
		//{NewJavBus, "abp"},
		//{NewMGStage, "277DCV"},
		{NewDMM, "ssis 13"},
	} {
		provider := unit.builder()
		results, err := provider.SearchMovie(unit.keyword)
		data, _ := json.MarshalIndent(results, "", "\t")
		assert.Nil(t, err)
		for _, result := range results {
			assert.True(t, result.Valid())
		}
		t.Logf("%s", data)
	}
}
