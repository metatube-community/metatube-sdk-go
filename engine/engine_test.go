package engine

import (
	"encoding/json"
	"testing"
	"time"

	_ "github.com/javtube/javtube-sdk-go/provider/1pondo"
	"github.com/stretchr/testify/assert"
)

func TestEngine_SearchMovieAll(t *testing.T) {
	engine := New(10 * time.Second)
	for _, item := range []string{
		"SSIS-033",
		"MIDV-003",
		"stars-138",
	} {
		results, err := engine.SearchMovieAll(item)
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
