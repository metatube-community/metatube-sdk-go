package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/internal/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func redirect(app *engine.Engine) gin.HandlerFunc {
	const queryKey = "redirect"

	return func(c *gin.Context) {
		if redir := c.Query(queryKey); redir != "" {
			var provider, id string
			if pid, err := providerid.Parse(redir); err != nil {
				abortWithStatusMessage(c, http.StatusBadRequest, err)
				return
			} else {
				provider, id = pid.Provider, pid.ID
			}

			var (
				info any
				err  error
			)
			switch {
			case app.IsActorProvider(provider):
				info, err = app.GetActorInfoByProviderID(provider, id, true)
			case app.IsMovieProvider(provider):
				info, err = app.GetMovieInfoByProviderID(provider, id, true)
			default:
				abortWithError(c, mt.ErrProviderNotFound)
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
