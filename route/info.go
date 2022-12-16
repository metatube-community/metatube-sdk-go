package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
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
	Lazy bool `form:"lazy"`
}

func getInfo(app *engine.Engine, typ infoType) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := &infoUri{}
		if err := c.ShouldBindUri(uri); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}
		query := &infoQuery{
			Lazy: true, // enable lazy by default.
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		var (
			info any
			err  error
		)
		switch typ {
		case actorInfoType:
			info, err = app.GetActorInfoByProviderID(uri.Provider, uri.ID, query.Lazy)
		case movieInfoType:
			info, err = app.GetMovieInfoByProviderID(uri.Provider, uri.ID, query.Lazy)
		default:
			panic("invalid info/metadata type")
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.JSON(http.StatusOK, &responseMessage{Data: info})
	}
}
