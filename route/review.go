package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/errors"
	"github.com/metatube-community/metatube-sdk-go/model"
)

type reviewQuery struct {
	Homepage string `form:"homepage"`
}

func getReview(app *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := &infoUri{}
		if err := c.ShouldBindUri(uri); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}
		query := &reviewQuery{}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		if !app.IsMovieProvider(uri.Provider) {
			abortWithError(c, errors.New(http.StatusBadRequest,
				"only movie provider is supported"))
			return
		}

		var (
			reviews []*model.MovieReviewInfo
			err     error
		)
		if query.Homepage != "" {
			reviews, err = app.GetMovieReviewsByProviderURL(uri.Provider, query.Homepage)
		} else {
			reviews, err = app.GetMovieReviewsByProviderID(uri.Provider, uri.ID)
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.JSON(http.StatusOK, &responseMessage{Data: reviews})
	}
}
