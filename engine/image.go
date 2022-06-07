package engine

import (
	"image"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/imageutil"
	"github.com/javtube/javtube-sdk-go/imageutil/pigo"
	R "github.com/javtube/javtube-sdk-go/internal/constant"
	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
)

// Default position constants for different kind of images.
const (
	defaultActorPrimaryImagePosition  = 0.5
	defaultMoviePrimaryImagePosition  = 1.0
	defaultMovieThumbImagePosition    = 0.5
	defaultMovieBackdropImagePosition = 0.0
)

func (e *Engine) GetActorPrimaryImage(id, name string) (image.Image, error) {
	info, err := e.GetActorInfoByProviderID(name, id, true)
	if err != nil {
		return nil, err
	}
	if len(info.Images) == 0 {
		return nil, javtube.ErrImageNotFound
	}
	return e.GetImageByURL(info.Images[0], e.MustGetActorProviderByName(name), R.PrimaryImageRatio, defaultActorPrimaryImagePosition, false)
}

func (e *Engine) GetMoviePrimaryImage(id, name string, pos float64) (image.Image, error) {
	url, info, err := e.getPreferredMovieImageURLAndInfo(id, name, true)
	if err != nil {
		return nil, err
	}
	var auto bool
	if pos < 0 /* manual position disabled */ {
		pos = defaultMoviePrimaryImagePosition
		auto = number.RequireFaceDetection(info.Number)
	}
	return e.GetImageByURL(url, e.MustGetMovieProviderByName(name), R.PrimaryImageRatio, pos, auto)
}

func (e *Engine) GetMovieThumbImage(id, name string) (image.Image, error) {
	url, _, err := e.getPreferredMovieImageURLAndInfo(id, name, false)
	if err != nil {
		return nil, err
	}
	return e.GetImageByURL(url, e.MustGetMovieProviderByName(name), R.ThumbImageRatio, defaultMovieThumbImagePosition, false)
}

func (e *Engine) GetMovieBackdropImage(id, name string) (image.Image, error) {
	url, _, err := e.getPreferredMovieImageURLAndInfo(id, name, false)
	if err != nil {
		return nil, err
	}
	return e.GetImageByURL(url, e.MustGetMovieProviderByName(name), R.BackdropImageRatio, defaultMovieBackdropImagePosition, false)
}

func (e *Engine) GetImageByURL(url string, provider javtube.Provider, ratio float64, pos float64, auto bool) (img image.Image, err error) {
	if img, err = e.getImageByURL(url, provider); err != nil {
		return
	}
	if auto {
		pos = pigo.CalculatePosition(img, ratio, pos)
	}
	return imageutil.CropImagePosition(img, ratio, pos), nil
}

func (e *Engine) getImageByURL(url string, provider javtube.Provider) (img image.Image, err error) {
	resp, err := e.Fetch(url, provider)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	img, _, err = image.Decode(resp.Body)
	return
}

func (e *Engine) getPreferredMovieImageURLAndInfo(id, name string, thumb bool) (url string, info *model.MovieInfo, err error) {
	info, err = e.GetMovieInfoByID(id, name, true)
	if err != nil {
		return
	}
	url = info.CoverURL
	if thumb && info.BigThumbURL != "" /* big thumb > cover */ {
		url = info.BigThumbURL
	} else if !thumb && info.BigCoverURL != "" /* big cover > cover */ {
		url = info.BigCoverURL
	}
	return
}
