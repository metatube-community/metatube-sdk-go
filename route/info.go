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

type infoUri struct {
	Provider string `uri:"provider" binding:"required"`
	ID       string `uri:"id" binding:"required"`
}

type infoQuery struct {
	URL  string `form:"url"`
	Lazy bool   `form:"lazy"`
}

func getInfo(app *engine.Engine, typ infoType) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := &infoUri{}
		if err := c.ShouldBindUri(uri); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}
		query := &infoQuery{
			Lazy: true,
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		if query.URL == "" && (uri.ID == "" || uri.Provider == "") {
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
				info, err = app.GetActorInfoByID(uri.ID, uri.Provider, query.Lazy)
			}
		case movieInfoType:
			if query.URL != "" {
				info, err = app.GetMovieInfoByURL(query.URL, query.Lazy)
			} else {
				info, err = app.GetMovieInfoByID(uri.ID, uri.Provider, query.Lazy)
			}
		default:
			panic("invalid info/metadata type")
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.PureJSON(http.StatusOK, &responseMessage{Data: info})
	}
}
