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
	ID       string `form:"id"`
	Provider string `form:"provider"`
	URL      string `form:"url"`
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
