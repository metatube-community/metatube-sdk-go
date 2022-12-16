package engine

import (
	"image"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	R "github.com/metatube-community/metatube-sdk-go/constant"
	"github.com/metatube-community/metatube-sdk-go/imageutil"
	"github.com/metatube-community/metatube-sdk-go/imageutil/pigo"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

// Default position constants for different kind of images.
const (
	defaultActorPrimaryImagePosition  = 0.5
	defaultMoviePrimaryImagePosition  = 1.0
	defaultMovieThumbImagePosition    = 0.5
	defaultMovieBackdropImagePosition = 0.0
)

func (e *Engine) GetActorPrimaryImage(name, id string) (image.Image, error) {
	info, err := e.GetActorInfoByProviderID(name, id, true)
	if err != nil {
		return nil, err
	}
	if len(info.Images) == 0 {
		return nil, mt.ErrImageNotFound
	}
	return e.GetImageByURL(e.MustGetActorProviderByName(name), info.Images[0], R.PrimaryImageRatio, defaultActorPrimaryImagePosition, false)
}

func (e *Engine) GetMoviePrimaryImage(name, id string, ratio, pos float64) (image.Image, error) {
	url, info, err := e.getPreferredMovieImageURLAndInfo(name, id, true)
	if err != nil {
		return nil, err
	}
	if ratio < 0 /* default primary ratio */ {
		ratio = R.PrimaryImageRatio
	}
	var auto bool
	if pos < 0 /* manual position disabled */ {
		pos = defaultMoviePrimaryImagePosition
		auto = number.RequireFaceDetection(info.Number)
	}
	return e.GetImageByURL(e.MustGetMovieProviderByName(name), url, ratio, pos, auto)
}

func (e *Engine) GetMovieThumbImage(name, id string) (image.Image, error) {
	url, _, err := e.getPreferredMovieImageURLAndInfo(name, id, false)
	if err != nil {
		return nil, err
	}
	return e.GetImageByURL(e.MustGetMovieProviderByName(name), url, R.ThumbImageRatio, defaultMovieThumbImagePosition, false)
}

func (e *Engine) GetMovieBackdropImage(name, id string) (image.Image, error) {
	url, _, err := e.getPreferredMovieImageURLAndInfo(name, id, false)
	if err != nil {
		return nil, err
	}
	return e.GetImageByURL(e.MustGetMovieProviderByName(name), url, R.BackdropImageRatio, defaultMovieBackdropImagePosition, false)
}

func (e *Engine) GetImageByURL(provider mt.Provider, url string, ratio, pos float64, auto bool) (img image.Image, err error) {
	if img, err = e.getImageByURL(provider, url); err != nil {
		return
	}
	if auto {
		pos = pigo.CalculatePosition(img, ratio, pos)
	}
	return imageutil.CropImagePosition(img, ratio, pos), nil
}

func (e *Engine) getImageByURL(provider mt.Provider, url string) (img image.Image, err error) {
	resp, err := e.Fetch(url, provider)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	img, _, err = image.Decode(resp.Body)
	return
}

func (e *Engine) getPreferredMovieImageURLAndInfo(name, id string, thumb bool) (url string, info *model.MovieInfo, err error) {
	info, err = e.GetMovieInfoByProviderID(name, id, true)
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
