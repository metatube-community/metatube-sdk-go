package modelmediaasia

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"
	"gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*ModelMediaAsia)(nil)
	_ provider.MovieSearcher = (*ModelMediaAsia)(nil)
)

const (
	Name     = "MODEL-MEDIA-ASIA"
	Priority = 0 // Disabled by default, use `export MT_MOVIE_PROVIDER_PRIORITY_MODEL_MEDIA_ASIA=1000` to enable.
)

const (
	baseURL      = "https://modelmediaasia.com/"
	movieURL     = "https://modelmediaasia.com/zh-CN/videos/%s"
	apiMovieURL  = "https://api.modelmediaasia.com/api/v2/videos/%s"
	apiSearchURL = "https://api.modelmediaasia.com/api/v2/search?keyword=%s"
)

type ModelMediaAsia struct {
	*scraper.Scraper
}

func New() *ModelMediaAsia {
	return &ModelMediaAsia{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Chinese)}
}

func (mma *ModelMediaAsia) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return mma.GetMovieInfoByURL(fmt.Sprintf(apiMovieURL, id))
}

func (mma *ModelMediaAsia) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

type movieInfoResponse struct {
	Data struct {
		ID            int    `json:"id"`
		SerialNumber  string `json:"serial_number"`
		Title         string `json:"title"`
		TitleCn       string `json:"title_cn"`
		Description   string `json:"description"`
		DescriptionCn string `json:"description_cn"`
		Trailer       string `json:"trailer"`
		Duration      int    `json:"duration"`
		Cover         string `json:"cover"`
		PreviewVideo  string `json:"preview_video"`
		PublishedAt   int64  `json:"published_at"`
		Models        []struct {
			ID                int    `json:"id"`
			Name              string `json:"name"`
			NameCn            string `json:"name_cn"`
			Avatar            string `json:"avatar"`
			Gender            string `json:"gender"`
			HeightFt          int    `json:"height_ft"`
			HeightIn          int    `json:"height_in"`
			WeightLbs         int    `json:"weight_lbs"`
			MeasurementsChest string `json:"measurements_chest"`
			MeasurementsWaist int    `json:"measurements_waist"`
			MeasurementsHips  int    `json:"measurements_hips"`

			// API return empty for following fields.
			Cover       string      `json:"cover"`
			MobileCover string      `json:"mobile_cover"`
			BirthDay    string      `json:"birth_day"`
			BirthPlace  string      `json:"birth_place"`
			HeightCm    int         `json:"height_cm"`
			WeightKg    int         `json:"weight_kg"`
			Videos      interface{} `json:"videos"`
			Photos      interface{} `json:"photos"`
		} `json:"models"`
		Tags []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			NameCn string `json:"name_cn"`
		} `json:"tags"`
	} `json:"data"`
}

func (mma *ModelMediaAsia) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := mma.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		Provider:      mma.Name(),
		Homepage:      fmt.Sprintf(movieURL, id),
		Maker:         "Model Media",
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

func (mma *ModelMediaAsia) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

type searchResponse struct {
	Data struct {
		Videos []struct {
			ID            int    `json:"id"`
			SerialNumber  string `json:"serial_number"`
			Title         string `json:"title"`
			TitleCn       string `json:"title_cn"`
			Description   string `json:"description"`
			DescriptionCn string `json:"description_cn"`
			Cover         string `json:"cover"`
			PublishedAt   int64  `json:"published_at"`
		} `json:"videos"`
	} `json:"data"`
}

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

// TODO: add support for actor search and scraping.

func init() {
	provider.Register(Name, New)
}
