package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/engine/providerid"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func redirect(app *engine.Engine) gin.HandlerFunc {
	const queryKey = "redirect"

	return func(c *gin.Context) {
		if redir := c.Query(queryKey); redir != "" {
			pid, err := providerid.Parse(redir)
			if err != nil {
				abortWithStatusMessage(c, http.StatusBadRequest, err)
				return
			}

			var info any
			switch {
			case app.IsActorProvider(pid.Provider):
				info, err = app.GetActorInfoByProviderID(pid, true)
			case app.IsMovieProvider(pid.Provider):
				info, err = app.GetMovieInfoByProviderID(pid, true)
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
