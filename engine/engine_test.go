package engine

import (
	"encoding/json"
	"testing"

	_ "github.com/javtube/javtube-sdk-go/provider/1pondo"
	"github.com/stretchr/testify/assert"
)

func TestEngine_SearchMovie(t *testing.T) {
	engine := New()
	for _, item := range []string{
		"SSIS-033",
		"MIDV-003",
		"stars-138",
	} {
		results, err := engine.SearchMovie(item)
		data, _ := json.MarshalIndent(results, "", "\t")
		if assert.NoError(t, err) {
			for _, result := range results {
				assert.True(t, result.Valid())
			}
		}
		t.Logf("%s", data)
	}
}

func TestEngine_GetMovieInfo(t *testing.T) {

}
