package modelmediaasia

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
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
	_ provider.ActorProvider = (*ModelMediaAsia)(nil)
	_ provider.ActorSearcher = (*ModelMediaAsia)(nil)
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
	*scraper.Scraper
}

func New() *ModelMediaAsia {
	return &ModelMediaAsia{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Chinese)}
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
		Models []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			NameCn string `json:"name_cn"`
			Avatar string `json:"avatar"`
		} `json:"models"`
	} `json:"data"`
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

type actorInfoResponse struct {
	Data struct {
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
		Cover       string `json:"cover"`
		MobileCover string `json:"mobile_cover"`
		Socialmedia string `json:"socialmedia"`
		PeriodViews int    `json:"period_views"`
		BirthDay    string `json:"birth_day"`
		BirthPlace  string `json:"birth_place"`
		Description string `json:"description"`
		Tooltips    string `json:"tooltips"`
		HeightCm    int    `json:"height_cm"`
		WeightKg    int    `json:"weight_kg"`
		Photos      []struct {
			Image string `json:"image"`
		} `json:"photos"`
	} `json:"data"`
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

		info.Name = resp.Data.NameCn
		if resp.Data.Name != "" {
			info.Aliases = append(info.Aliases, resp.Data.Name)
		}

		info.Images = append(info.Images, resp.Data.Avatar)
		for _, photo := range resp.Data.Photos {
			info.Images = append(info.Images, photo.Image)
		}

		info.Height = lengthConversion(resp.Data.HeightFt, resp.Data.HeightIn)

		parseChestSize(resp.Data.MeasurementsChest)
		b, cup, err := parseChestSize(resp.Data.MeasurementsChest)
		if err == nil {
			info.CupSize = cup
			if b != 0 && resp.Data.MeasurementsWaist != 0 && resp.Data.MeasurementsHips != 0 {
				info.Measurements = fmt.Sprintf("B:%d / W:%d / H:%d", b, resp.Data.MeasurementsWaist, resp.Data.MeasurementsHips)
			}
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

	results = make([]*model.ActorSearchResult, 0)

	c.OnResponse(func(r *colly.Response) {
		resp := &searchResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}
		for _, actor := range resp.Data.Models {
			res := &model.ActorSearchResult{
				ID:       strconv.Itoa(actor.ID),
				Name:     actor.NameCn,
				Provider: mma.Name(),
				Homepage: fmt.Sprintf(actorURL, strconv.Itoa(actor.ID)),
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

// lengthConversion converts feet + inch to cm.
func lengthConversion(feet, inch int) int {
	return int(float32(feet*12+inch) * 2.54)
}

var chestSizeRE = regexp.MustCompile(`^(\d+)([A-Z])$`)

func parseChestSize(s string) (int, string, error) {
	match := chestSizeRE.FindStringSubmatch(s)

	if len(match) != 3 {
		return 0, "", fmt.Errorf("invalid format: %s", s)
	}

	numericPart := match[1]
	unitPart := match[2]

	value, err := strconv.Atoi(numericPart)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse numeric part '%s': %w", numericPart, err)
	}

	return value, unitPart, nil
}
