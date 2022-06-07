package route

import (
	"net/http"
	pkgurl "net/url"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/model"
)

type searchType uint8

const (
	actorSearchType searchType = iota
	movieSearchType
)

type searchQuery struct {
	Q        string `form:"q" binding:"required"`
	Provider string `form:"provider"`
	Lazy     bool   `form:"lazy"`
}

func getSearch(app *engine.Engine, typ searchType) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := &searchQuery{
			Lazy: false, // disable lazy by default.
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		isValidURL := true
		if _, err := pkgurl.ParseRequestURI(query.Q); err != nil {
			isValidURL = false
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
			if isValidURL {
				results, err = app.GetActorInfoByURL(query.Q, true /* always lazy */)
			} else if searchAll {
				results, err = app.SearchActorAll(query.Q)
			} else {
				results, err = app.SearchActor(query.Q, query.Provider, query.Lazy)
			}
		case movieSearchType:
			if isValidURL {
				results, err = app.GetMovieInfoByURL(query.Q, true /* always lazy */)
			} else if searchAll {
				results, err = app.SearchMovieAll(query.Q, query.Lazy)
			} else {
				results, err = app.SearchMovie(query.Q, query.Provider, query.Lazy)
			}
		default:
			panic("invalid search type")
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		// convert to search results.
		switch v := results.(type) {
		case *model.ActorInfo:
			results = []*model.ActorSearchResult{v.ToSearchResult()}
		case *model.MovieInfo:
			results = []*model.MovieSearchResult{v.ToSearchResult()}
		}

		c.PureJSON(http.StatusOK, &responseMessage{Data: results})
	}
}
