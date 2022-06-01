package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/model"
)

func redirect(app *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		if url := c.Query("redirect"); url != "" {
			var (
				provider string
				id       string
			)
			if ss := strings.Split(url, ":"); len(ss) > 1 {
				provider, id = ss[0], ss[1]
			}

			var (
				info any
				err  error
			)
			switch {
			case app.IsActorProvider(provider):
				info, err = app.GetActorInfoByID(id, provider, true)
			case app.IsMovieProvider(provider):
				info, err = app.GetMovieInfoByID(id, provider, true)
			default:
				abortWithStatusMessage(c, http.StatusBadRequest, "invalid provider")
				return
			}
			if err != nil {
				abortWithError(c, err)
				return
			}

			var homepage string
			switch v := info.(type) {
			case *model.ActorInfo:
				homepage = v.Homepage
			case *model.MovieInfo:
				homepage = v.Homepage
			}
			c.Redirect(http.StatusTemporaryRedirect, homepage)

			c.Abort() // abort pending middlewares
			return
		}
		c.Next()
	}
}
