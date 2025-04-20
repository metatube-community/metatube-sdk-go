package route

import (
	"bytes"
	"image"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"

	R "github.com/metatube-community/metatube-sdk-go/constant"
	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/imageutil"
	"github.com/metatube-community/metatube-sdk-go/imageutil/badge"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
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
	Badge    string  `form:"badge"`
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
			Quality:  90,
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		// TODO: how to handle providers that implement
		//   both actor and movie provider interfaces?
		var isActorProvider bool
		switch {
		case app.IsActorProvider(uri.Provider):
			isActorProvider = true
		case app.IsMovieProvider(uri.Provider):
			isActorProvider = false
		default:
			abortWithError(c, mt.ErrProviderNotFound)
			return
		}

		var (
			img image.Image
			err error
		)
		if query.URL != "" /* specified URL */ {
			var provider mt.Provider
			if isActorProvider {
				provider = app.MustGetActorProviderByName(uri.Provider)
			} else {
				provider = app.MustGetMovieProviderByName(uri.Provider)
			}
			// query.Ratio should apply only to the primary images.
			if typ != primaryImageType || query.Ratio < 0 {
				query.Ratio = ratio
			}
			img, err = app.GetImageByURL(provider, query.URL, query.Ratio, query.Position, query.Auto)
		} else if isActorProvider /* actor */ {
			switch typ {
			case primaryImageType:
				img, err = app.GetActorPrimaryImage(uri.AsProviderID())
			case thumbImageType, backdropImageType:
				abortWithStatusMessage(c, http.StatusBadRequest, "unsupported image type")
				return
			}
		} else /* movie */ {
			switch typ {
			case primaryImageType:
				img, err = app.GetMoviePrimaryImage(uri.AsProviderID(), query.Ratio, query.Position)
			case thumbImageType:
				img, err = app.GetMovieThumbImage(uri.AsProviderID())
			case backdropImageType:
				img, err = app.GetMovieBackdropImage(uri.AsProviderID())
			}
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		if query.Badge != "" {
			if img, err = badge.Badge(img, query.Badge); err != nil {
				abortWithError(c, err)
				return
			}
		}

		c.Header("X-MetaTube-Image-Width", strconv.Itoa(img.Bounds().Dx()))
		c.Header("X-MetaTube-Image-Height", strconv.Itoa(img.Bounds().Dy()))

		buf := &bytes.Buffer{}
		if err = imageutil.EncodeToJPEG(buf, img, query.Quality); err != nil {
			panic(err)
		}

		c.Render(http.StatusOK, render.Reader{
			ContentType:   jpegImageMIMEType,
			ContentLength: int64(buf.Len()),
			Reader:        buf,
		})
	}
}

const jpegImageMIMEType = "image/jpeg"
