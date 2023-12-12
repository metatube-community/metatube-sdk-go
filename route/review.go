package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/errors"
)

func getReview(app *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := &infoUri{}
		if err := c.ShouldBindUri(uri); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		if !app.IsMovieProvider(uri.Provider) {
			abortWithError(c, errors.New(http.StatusBadRequest,
				"only movie provider is supported"))
			return
		}

		reviews, err := app.GetMovieReviewInfoByProviderID(uri.Provider, uri.ID)
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.JSON(http.StatusOK, &responseMessage{Data: reviews})
	}
}
