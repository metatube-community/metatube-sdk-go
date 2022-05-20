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
	ID       string `form:"id" binding:"required"`
	Provider string `form:"provider" binding:"required"`
	Update   bool   `form:"update"`
}

func GetInfo(app *engine.Engine, typ infoType) gin.HandlerFunc {
	return func(c *gin.Context) {
		var query infoQuery
		if err := c.ShouldBindQuery(&query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		var (
			info any
			err  error
		)
		switch typ {
		case actorInfoType:
			info, err = app.GetActorInfoByID(query.ID, query.Provider, !query.Update)
		case movieInfoType:
			info, err = app.GetMovieInfoByID(query.ID, query.Provider, !query.Update)
		default:
			panic("invalid info/metadata type")
		}
		if err != nil {
			abortWithStatusMessage(c, http.StatusInternalServerError, err)
			return
		}

		// JSON reply.
		c.PureJSON(http.StatusOK, info)
	}
}
