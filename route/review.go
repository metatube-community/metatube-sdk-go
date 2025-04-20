package route

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/errors"
	"github.com/metatube-community/metatube-sdk-go/model"
)

type reviewUri struct {
	infoUri // same as info uri
}

type reviewQuery struct {
	Homepage string `form:"homepage"`
	Lazy     bool   `form:"lazy"`
}

func getReview(app *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := &reviewUri{}
		if err := c.ShouldBindUri(uri); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}
		query := &reviewQuery{
			Lazy: true, // enable lazy by default.
		}
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
			reviews *model.MovieReviewInfo
			err     error
		)
		if query.Homepage != "" {
			reviews, err = app.GetMovieReviewsByProviderURL(query.Homepage, query.Lazy)
		} else {
			reviews, err = app.GetMovieReviewsByProviderID(uri.AsProviderID(), query.Lazy)
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.JSON(http.StatusOK, &responseMessage{Data: reviews.Reviews.Data()})
	}
}
