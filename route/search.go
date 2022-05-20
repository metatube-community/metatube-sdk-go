package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javtube/javtube-sdk-go/engine"
)

type searchType uint8

const (
	actorSearchType searchType = iota
	movieSearchType
)

type searchQuery struct {
	Keyword  string `form:"keyword" binding:"required"`
	Provider string `form:"provider"`
	Update   bool   `form:"update"`
}

func getSearchResult(app *engine.Engine, typ searchType) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query searchQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		fuzzSearch := true
		if query.Provider != "" {
			fuzzSearch = false
		}

		var (
			results any
			err     error
		)
		switch typ {
		case actorSearchType:
			if fuzzSearch {
				results, err = app.SearchActorAll(query.Keyword)
			} else {
				results, err = app.SearchActor(query.Keyword, query.Provider, !query.Update)
			}
		case movieSearchType:
			if fuzzSearch {
				results, err = app.SearchMovieAll(query.Keyword, !query.Update)
			} else {
				results, err = app.SearchMovie(query.Keyword, query.Provider, !query.Update)
			}
		default:
			panic("invalid search type")
		}
		if err != nil {
			abortWithStatusMessage(c, http.StatusInternalServerError, err)
			return
		}

		// JSON reply.
		c.PureJSON(http.StatusOK, results)
	}
}
