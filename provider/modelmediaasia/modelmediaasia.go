package modelmediaasia

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"
	"gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/collection/sets"
	"github.com/metatube-community/metatube-sdk-go/common/convertor"
	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.ActorProvider = (*ModelMediaAsia)(nil)
	_ provider.ActorSearcher = (*ModelMediaAsia)(nil)
	_ provider.MovieProvider = (*ModelMediaAsia)(nil)
	_ provider.MovieSearcher = (*ModelMediaAsia)(nil)
	_ provider.Fetcher       = (*ModelMediaAsia)(nil)
)

const (
	Name     = "ModelMediaAsia"
	Priority = 0 // Disabled by default, use `export MT_PROVIDER_MODELMEDIAASIA__PRIORITY=1000` to enable.
)

const (
	baseURL      = "https://modelmediaasia.com/"
	movieURL     = "https://modelmediaasia.com/zh-CN/videos/%s"
	actorURL     = "https://modelmediaasia.com/zh-CN/models/%s"
	apiMovieURL  = "https://api.modelmediaasia.com/api/v2/videos/%s"
	apiActorURL  = "https://api.modelmediaasia.com/api/v2/models/%s"
	apiSearchURL = "https://api.modelmediaasia.com/api/v2/search?keyword=%s"
)

type ModelMediaAsia struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func New() *ModelMediaAsia {
	return &ModelMediaAsia{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL}),
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Chinese),
	}
}

// GetMovieInfoByID impls MovieProvider.GetMovieInfoByID.
func (mma *ModelMediaAsia) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	info = &model.MovieInfo{
		Provider:      mma.Name(),
		Homepage:      fmt.Sprintf(movieURL, id),
		Maker:         "麻豆傳媒映畫",
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := mma.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		resp := &movieInfoResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}

		info.ID = resp.Data.SerialNumber
		info.Number = resp.Data.SerialNumber
		info.Title = resp.Data.TitleCn
		info.Summary = resp.Data.DescriptionCn
		info.ThumbURL = resp.Data.Cover
		info.CoverURL = resp.Data.Cover
		info.ReleaseDate = datatypes.Date(time.UnixMilli(resp.Data.PublishedAt))

		// Trailer > PreviewVideo
		info.PreviewVideoURL = map[bool]string{
			true:  resp.Data.Trailer,
			false: resp.Data.PreviewVideo,
		}[resp.Data.Trailer != ""]

		for _, tag := range resp.Data.Tags {
			info.Genres = append(info.Genres, tag.NameCn)
		}

		for _, actor := range resp.Data.Models {
			info.Actors = append(info.Actors, actor.NameCn)
		}
	})

	err = c.Visit(fmt.Sprintf(apiMovieURL, id))
	return
}

// ParseMovieIDFromURL impls MovieProvider.ParseMovieIDFromURL.
func (mma *ModelMediaAsia) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

// GetMovieInfoByURL impls MovieProvider.GetMovieInfoByURL.
func (mma *ModelMediaAsia) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := mma.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	return mma.GetMovieInfoByID(id)
}

// NormalizeMovieKeyword impls MovieSearcher.NormalizeMovieKeyword.
func (mma *ModelMediaAsia) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

// SearchMovie impls MovieSearcher.SearchMovie.
func (mma *ModelMediaAsia) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := mma.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		resp := &searchResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}
		for _, video := range resp.Data.Videos {
			results = append(results, &model.MovieSearchResult{
				ID:          video.SerialNumber,
				Number:      video.SerialNumber,
				Title:       video.TitleCn,
				Provider:    mma.Name(),
				Homepage:    fmt.Sprintf(movieURL, video.SerialNumber),
				ThumbURL:    video.Cover,
				CoverURL:    video.Cover,
				ReleaseDate: datatypes.Date(time.UnixMilli(video.PublishedAt)),
			})
		}
	})

	err = c.Visit(fmt.Sprintf(apiSearchURL, url.QueryEscape(keyword)))
	return
}

// ParseActorIDFromURL impls ActorProvider.ParseActorIDFromURL.
func (mma *ModelMediaAsia) ParseActorIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

// GetActorInfoByID impls ActorProvider.GetActorInfoByID.
func (mma *ModelMediaAsia) GetActorInfoByID(id string) (info *model.ActorInfo, err error) {
	info = &model.ActorInfo{
		ID:       id,
		Provider: mma.Name(),
		Homepage: fmt.Sprintf(actorURL, id),
		Aliases:  []string{},
		Images:   []string{},
	}

	c := mma.ClonedCollector()
	c.OnResponse(func(r *colly.Response) {
		resp := &actorInfoResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}

		// Name & Aliases
		info.Name = resp.Data.NameCn

		if resp.Data.Name != "" /* English name */ {
			info.Aliases = append(info.Aliases, resp.Data.Name)
		}

		// Images
		imageSet := sets.NewOrderedSet[string]()
		imageSet.Add(resp.Data.Avatar)
		for _, photo := range resp.Data.Photos {
			imageSet.Add(photo.Image)
		}
		info.Images = imageSet.AsSlice()

		// Birthday
		info.Birthday = parser.ParseDate(resp.Data.BirthDay)

		// Height
		if resp.Data.HeightCm > 0 {
			info.Height = resp.Data.HeightCm
		} else {
			info.Height = convertor.ConvertToCentimeters(
				resp.Data.HeightFt, resp.Data.HeightIn)
		}

		if size, cup, err := parser.ParseBustCupSize(resp.Data.MeasurementsChest); err == nil {
			if size != 0 &&
				resp.Data.MeasurementsWaist != 0 &&
				resp.Data.MeasurementsHips != 0 {
				info.Measurements = fmt.Sprintf("B:%d / W:%d / H:%d",
					size, resp.Data.MeasurementsWaist, resp.Data.MeasurementsHips)
			}
			info.CupSize = cup
		}
	})

	err = c.Visit(fmt.Sprintf(apiActorURL, id))
	return
}

// GetActorInfoByURL impls ActorProvider.GetActorInfoByURL.
func (mma *ModelMediaAsia) GetActorInfoByURL(rawURL string) (*model.ActorInfo, error) {
	id, err := mma.ParseActorIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}

	return mma.GetActorInfoByID(id)
}

// SearchActor impls ActorSearcher.SearchActor.
func (mma *ModelMediaAsia) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	c := mma.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		resp := &searchResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}
		for _, actor := range resp.Data.Models {
			actorID := strconv.Itoa(actor.ID)
			res := &model.ActorSearchResult{
				ID:       actorID,
				Name:     actor.NameCn,
				Provider: mma.Name(),
				Homepage: fmt.Sprintf(actorURL, actorID),
			}
			if actor.Avatar != "" {
				res.Images = append(res.Images, actor.Avatar)
			}
			if actor.Name != "" {
				res.Aliases = append(res.Aliases, actor.Name)
			}
			results = append(results, res)
		}
	})

	err = c.Visit(fmt.Sprintf(apiSearchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.Register(Name, New)
}
