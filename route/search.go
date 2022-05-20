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
	Lazy     bool   `form:"lazy"`
}

func getSearchResults(app *engine.Engine, typ searchType) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := &searchQuery{
			Lazy: false, // disable lazy by default.
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		searchAll := true
		if query.Provider != "" {
			searchAll = false
		}

		var (
			results any
			err     error
		)
		switch typ {
		case actorSearchType:
			if searchAll {
				results, err = app.SearchActorAll(query.Keyword)
			} else {
				results, err = app.SearchActor(query.Keyword, query.Provider, query.Lazy)
			}
		case movieSearchType:
			if searchAll {
				results, err = app.SearchMovieAll(query.Keyword, query.Lazy)
			} else {
				results, err = app.SearchMovie(query.Keyword, query.Provider, query.Lazy)
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
