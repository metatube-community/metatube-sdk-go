package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/translate"
)

type infoType uint8

const (
	actorInfoType infoType = iota
	movieInfoType
)

type infoQuery struct {
	ID       string `form:"id" binding:"required"`
	Provider string `form:"provider" binding:"required"`
	Language string `form:"lang"`
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

		var (
			info any
			err  error
		)
		switch typ {
		case actorInfoType:
			info, err = app.GetActorInfoByID(query.ID, query.Provider, query.Lazy)
		case movieInfoType:
			info, err = app.GetMovieInfoByID(query.ID, query.Provider, query.Lazy)
		default:
			panic("invalid info/metadata type")
		}
		if err != nil {
			abortWithStatusMessage(c, http.StatusInternalServerError, err)
			return
		}

		// translate
		if typ == movieInfoType && query.Language != "" {
			info := info.(*model.MovieInfo)
			if info.Title != "" {
				if title, _ := translate.Translate(info.Title, "auto", query.Language); title != "" {
					info.Title = title
				}
			}
			if info.Summary != "" {
				if summary, _ := translate.Translate(info.Summary, "auto", query.Language); summary != "" {
					info.Summary = summary
				}
			}
		}

		// JSON reply.
		c.PureJSON(http.StatusOK, info)
	}
}
