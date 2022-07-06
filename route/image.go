package route

import (
	"bytes"
	"image"
	"image/jpeg"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	R "github.com/javtube/javtube-sdk-go/constant"
	"github.com/javtube/javtube-sdk-go/engine"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

type imageType uint8

const (
	primaryImageType imageType = iota
	thumbImageType
	backdropImageType
)

type imageUri struct {
	infoUri // same as info uri
}

type imageQuery struct {
	URL      string  `form:"url"`
	Ratio    float64 `form:"ratio"`
	Position float64 `form:"pos"`
	Auto     bool    `form:"auto"`
	Quality  int     `form:"quality"`
}

func getImage(app *engine.Engine, typ imageType) gin.HandlerFunc {
	var ratio float64
	switch typ {
	case primaryImageType:
		ratio = R.PrimaryImageRatio
	case thumbImageType:
		ratio = R.ThumbImageRatio
	case backdropImageType:
		ratio = R.BackdropImageRatio
	default:
		panic("invalid image type")
	}

	return func(c *gin.Context) {
		uri := &imageUri{}
		if err := c.ShouldBindUri(uri); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}
		query := &imageQuery{
			Ratio:    -1,
			Position: -1,
			Quality:  95,
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		var isActorProvider bool
		switch {
		case app.IsActorProvider(uri.Provider):
			isActorProvider = true
		case app.IsMovieProvider(uri.Provider):
			isActorProvider = false
		default:
			abortWithError(c, javtube.ErrProviderNotFound)
			return
		}

		var (
			img image.Image
			err error
		)
		if query.URL != "" /* specified URL */ {
			var provider javtube.Provider
			if isActorProvider {
				provider = app.MustGetActorProviderByName(uri.Provider)
			} else {
				provider = app.MustGetMovieProviderByName(uri.Provider)
			}
			img, err = app.GetImageByURL(provider, query.URL, ratio, query.Position, query.Auto)
		} else if isActorProvider /* actor */ {
			switch typ {
			case primaryImageType:
				img, err = app.GetActorPrimaryImage(uri.Provider, uri.ID)
			case thumbImageType, backdropImageType:
				abortWithStatusMessage(c, http.StatusBadRequest, "unsupported image type")
				return
			}
		} else /* movie */ {
			switch typ {
			case primaryImageType:
				img, err = app.GetMoviePrimaryImage(uri.Provider, uri.ID, query.Ratio, query.Position)
			case thumbImageType:
				img, err = app.GetMovieThumbImage(uri.Provider, uri.ID)
			case backdropImageType:
				img, err = app.GetMovieBackdropImage(uri.Provider, uri.ID)
			}
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.Header("X-JavTube-Image-Width", strconv.Itoa(img.Bounds().Dx()))
		c.Header("X-JavTube-Image-Height", strconv.Itoa(img.Bounds().Dy()))

		buf := &bytes.Buffer{}
		if err = jpeg.Encode(buf, img, &jpeg.Options{Quality: query.Quality}); err != nil {
			panic(err)
		}

		c.Render(http.StatusOK, render.Reader{
			ContentType:   jpegImageMIMEType,
			ContentLength: int64(buf.Len()),
			Reader:        buf,
			Headers: map[string]string{
				// should be cached for a week.
				"Cache-Control": "max-age=604800, public",
			},
		})
	}
}

const jpegImageMIMEType = "image/jpeg"
