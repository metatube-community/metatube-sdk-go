package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/engine"
)

type infoType uint8

const (
	actorInfoType infoType = iota
	movieInfoType
)

type infoQuery struct {
	ID       string `form:"id"`
	Provider string `form:"provider"`
	URL      string `form:"url"`
	Lazy     bool   `form:"lazy"`
}

func getInfo(app *engine.Engine, typ infoType) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := &infoQuery{
			Lazy: true,
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		if query.URL == "" && (query.ID == "" || query.Provider == "") {
			abortWithStatusMessage(c, http.StatusBadRequest, "bad query")
			return
		}

		var (
			info any
			err  error
		)
		switch typ {
		case actorInfoType:
			if query.URL != "" {
				info, err = app.GetActorInfoByURL(query.URL, query.Lazy)
			} else {
				info, err = app.GetActorInfoByID(query.ID, query.Provider, query.Lazy)
			}
		case movieInfoType:
			if query.URL != "" {
				info, err = app.GetMovieInfoByURL(query.URL, query.Lazy)
			} else {
				info, err = app.GetMovieInfoByID(query.ID, query.Provider, query.Lazy)
			}
		default:
			panic("invalid info/metadata type")
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		// JSON reply.
		c.JSON(http.StatusOK, &responseMessage{
			Success: true,
			Data:    info,
		})
	}
}
