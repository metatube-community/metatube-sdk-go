package route

import (
	"errors"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/internal/define"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

type imageType uint8

const (
	primaryImageType imageType = iota
	thumbImageType
	backdropImageType
)

type imageQuery struct {
	ID       string  `form:"id" binding:"required"`
	Provider string  `form:"provider" binding:"required"`
	URL      string  `form:"url"`
	Position float64 `form:"pos"`
	Auto     bool    `form:"auto"`
	Quality  int     `form:"quality"`
}

func getImage(app *engine.Engine, typ imageType) gin.HandlerFunc {
	var ratio float64
	switch typ {
	case primaryImageType:
		ratio = define.PrimaryImageRatio
	case thumbImageType:
		ratio = define.ThumbImageRatio
	case backdropImageType:
		ratio = define.BackdropImageRatio
	default:
		panic("invalid image type")
	}

	return func(c *gin.Context) {
		query := &imageQuery{
			Position: -1,
			Quality:  95,
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		var isActorProvider bool
		switch {
		case app.IsActorProvider(query.Provider):
			isActorProvider = true
		case app.IsMovieProvider(query.Provider):
			isActorProvider = false
		default:
			abortWithStatusMessage(c, http.StatusBadRequest, "invalid provider")
			return
		}

		var (
			img  image.Image
			err  error
			code = http.StatusInternalServerError
		)
		if query.URL != "" /* specified URL */ {
			var provider javtube.Provider
			if isActorProvider {
				provider = app.MustGetActorProvider(query.Provider)
			} else {
				provider = app.MustGetMovieProvider(query.Provider)
			}
			img, err = app.GetImageByURL(query.URL, provider, ratio, query.Position, query.Auto)
		} else if isActorProvider /* actor */ {
			switch typ {
			case primaryImageType:
				img, err = app.GetActorPrimaryImage(query.ID, query.Provider)
			case thumbImageType, backdropImageType:
				code = http.StatusBadRequest
				err = errors.New("unsupported image type")
			}
		} else /* movie */ {
			switch typ {
			case primaryImageType:
				img, err = app.GetMoviePrimaryImage(query.ID, query.Provider, query.Position)
			case thumbImageType:
				img, err = app.GetMovieThumbImage(query.ID, query.Provider)
			case backdropImageType:
				img, err = app.GetMovieBackdropImage(query.ID, query.Provider)
			}
		}

		if err != nil {
			abortWithStatusMessage(c, code, err)
			return
		}

		c.Header("Content-Type", "image/jpeg")
		_ = jpeg.Encode(c.Writer, img, &jpeg.Options{Quality: query.Quality})
	}
}
