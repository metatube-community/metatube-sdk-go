package route

import (
	"image"
	"image/jpeg"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javtube/javtube-sdk-go/engine"
)

type imageType uint8

const (
	primaryImageType imageType = iota
	thumbImageType
	backdropImageType
)

type imageQuery struct {
	ID       string  `form:"id"`
	Provider string  `form:"provider"`
	URL      string  `form:"url"`
	Ratio    float64 `form:"ratio"`
	Position float64 `form:"pos"`
	Quality  int     `form:"quality"`
}

func getImage(app *engine.Engine, typ imageType) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := &imageQuery{
			Quality: 90,
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		var (
			img image.Image
			err error
		)
		switch typ {
		case primaryImageType:
			img, err = app.GetMoviePrimaryImage(query.ID, query.Provider)
		case thumbImageType:
			img, err = app.GetMovieThumbImage(query.ID, query.Provider)
		case backdropImageType:
			img, err = app.GetMovieBackdropImage(query.ID, query.Provider)
		}
		if err != nil {
			c.Error(err)
			return
		}

		c.Header("Content-Type", "image/jpeg")
		_ = jpeg.Encode(c.Writer, img, &jpeg.Options{Quality: query.Quality})
	}
}
