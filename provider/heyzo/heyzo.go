package heyzo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/convertor"
	"github.com/metatube-community/metatube-sdk-go/common/m3u8"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*Heyzo)(nil)
	_ provider.MovieReviewer = (*Heyzo)(nil)
)

const (
	Name     = "HEYZO"
	Priority = 1000
)

const (
	baseURL        = "https://www.heyzo.com/"
	movieURL       = "https://www.heyzo.com/movie/%04s"
	movieAPIURL    = "https://www.heyzo.com/api/v1/movie?movieId=%04s"
	reviewAPIURL   = "https://www.heyzo.com/api/v1/movie/%d/review?page=%d&limit=%d"
	reviewPageSize = 100
)

// Media URLs, built from the movie ID the same way as the official site
// frontend does, where `3000` is the HEYZO site ID.
const (
	coverURL     = "https://www.heyzo.com/contents/3000/%s/images/player_thumbnail.jpg"
	galleryURL   = "https://www.heyzo.com/contents/3000/%s/gallery/%03d.jpg"
	sampleHLSURL = "https://hls.heyzo.com/sample/3000/%s/mb.m3u8"
	sampleURL    = "https://www.heyzo.com/contents/%s/%s/%s"
	galleryCount = 21 // fixed gallery size of every movie.
)

// Category levels used by HEYZO.
const (
	categoryLevelFlag   = 3  // internal site flags, e.g., `60p`, `release preview`.
	categoryLevelSeries = 10 // series.
)

// Movie attribute IDs used by HEYZO.
const (
	attributeKeywords = 1  // space separated keywords.
	attributeDuration = 19 // JSON encoded durations, e.g., {"full":"01:01:30"}.
)

var errMovieNotFound = errors.New("movie not found")

type Heyzo struct {
	*scraper.Scraper
}

func New() *Heyzo {
	return &Heyzo{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (hzo *Heyzo) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:heyzo[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (hzo *Heyzo) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// Supports both new `/movie/0841` and legacy `/moviepages/0841/index.html` URLs.
	if ss := regexp.MustCompile(`/movie(?:pages)?/(\d+)`).FindStringSubmatch(homepage.Path); len(ss) == 2 {
		return ss[1], nil
	}
	return "", fmt.Errorf("invalid %s movie URL: %s", Name, rawURL)
}

func (hzo *Heyzo) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return hzo.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (hzo *Heyzo) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := hzo.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	data, err := hzo.fetchMovie(id)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            data.MovieID,
		Number:        fmt.Sprintf("HEYZO-%s", data.MovieID),
		Title:         data.Name,
		Summary:       convertor.NormalizeNewlines(data.Detail.Memo),
		Provider:      hzo.Name(),
		Homepage:      fmt.Sprintf(movieURL, data.MovieID),
		Maker:         "HEYZO",
		Score:         data.Detail.RvScore,
		ReleaseDate:   parser.ParseDate(data.MonthlyStartDate),
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	// Actors
	for _, actor := range data.Actors {
		if name := strings.TrimSpace(actor.ActorName); name != "" {
			info.Actors = append(info.Actors, name)
		}
	}

	// Genres+Series
	genreSet := make(map[string]struct{})
	addGenre := func(genre string) {
		if genre = strings.TrimSpace(genre); genre == "" {
			return
		}
		if _, ok := genreSet[genre]; ok {
			return
		}
		genreSet[genre] = struct{}{}
		info.Genres = append(info.Genres, genre)
	}
	for _, category := range data.Categories {
		switch category.Level {
		case categoryLevelFlag:
			// skip internal flags.
		case categoryLevelSeries:
			if info.Series == "" {
				info.Series = strings.TrimSpace(category.CategoryName)
			}
		default:
			addGenre(category.CategoryName)
		}
	}

	// Attributes
	for _, attr := range data.Attributes {
		switch attr.AttributeID {
		case attributeKeywords:
			for _, keyword := range strings.Fields(attr.AttributeValue) {
				addGenre(keyword)
			}
		case attributeDuration:
			duration := movieDuration{}
			if json.Unmarshal([]byte(attr.AttributeValue), &duration) == nil {
				info.Runtime = parser.ParseRuntime(duration.Full)
			}
		}
	}
	// Runtime (fallback)
	if info.Runtime == 0 {
		for _, file := range data.Download.Whole {
			if runtime := parser.ParseRuntime(file.Duration); runtime > 0 {
				info.Runtime = runtime
				break
			}
		}
	}

	// Cover+Thumb
	info.CoverURL = fmt.Sprintf(coverURL, data.MovieID)
	info.ThumbURL = info.CoverURL /* use cover as thumb */

	// Preview Images (the site shows no gallery for bonus/limited movies)
	if !data.IsBonus && !data.IsLimited && len(data.Actors) > 0 {
		for i := 1; i <= galleryCount; i++ {
			info.PreviewImages = append(info.PreviewImages, fmt.Sprintf(galleryURL, data.MovieID, i))
		}
	}

	// Preview Video
	hzo.fetchPreviewVideo(info)
	return
}

// fetchPreviewVideo resolves the sample video URLs from the HLS playlist.
// It is best-effort, as not every movie provides a sample video.
func (hzo *Heyzo) fetchPreviewVideo(info *model.MovieInfo) {
	c := hzo.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		info.PreviewVideoHLSURL = r.Request.URL.String()
		if uri, _, err := m3u8.ParseBestMediaURI(bytes.NewReader(r.Body)); err == nil {
			if ss := regexp.MustCompile(`/sample/(\d+)/(\d+)/ts\.(.+?)\.m3u8`).
				FindStringSubmatch(uri); len(ss) == 4 {
				info.PreviewVideoURL = fmt.Sprintf(sampleURL, ss[1], ss[2], ss[3])
			}
		}
	})

	c.Visit(fmt.Sprintf(sampleHLSURL, info.ID))
}

func (hzo *Heyzo) GetMovieReviewsByID(id string) (reviews []*model.MovieReviewDetail, err error) {
	data, err := hzo.fetchMovie(id)
	if err != nil {
		return
	}

	c := hzo.ClonedCollector()

	var (
		total int
		count int
	)
	c.OnResponse(func(r *colly.Response) {
		resp := reviewResponse{}
		if err = json.Unmarshal(r.Body, &resp); err != nil {
			return
		}
		total = resp.TotalCount
		count = len(resp.Items)
		for _, item := range resp.Items {
			comment := convertor.NormalizeNewlines(item.Comment)
			if item.Username == "" || comment == "" {
				continue
			}
			reviews = append(reviews, &model.MovieReviewDetail{
				Author:  item.Username,
				Comment: comment,
				Score:   item.Score,
				Date:    parser.ParseDate(item.Created),
			})
		}
	})

	for page, fetched := 1, 0; ; page++ {
		count = 0
		if vErr := c.Visit(fmt.Sprintf(reviewAPIURL, data.MovieSeq, page, reviewPageSize)); vErr != nil {
			err = vErr
			return
		}
		if err != nil {
			return
		}
		if fetched += count; count == 0 || fetched >= total {
			break
		}
	}
	return
}

func (hzo *Heyzo) GetMovieReviewsByURL(rawURL string) (reviews []*model.MovieReviewDetail, err error) {
	id, err := hzo.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}
	return hzo.GetMovieReviewsByID(id)
}

func (hzo *Heyzo) fetchMovie(id string) (data *movieResponse, err error) {
	c := hzo.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		data = &movieResponse{}
		if err = json.Unmarshal(r.Body, data); err != nil {
			return
		}
		if data.MovieSeq == 0 || data.MovieID == "" {
			err = errMovieNotFound
		}
	})

	if vErr := c.Visit(fmt.Sprintf(movieAPIURL, url.QueryEscape(id))); vErr != nil {
		if vErr.Error() == http.StatusText(http.StatusNotFound) {
			// API responds 404 with {"error":"Movie not found"}.
			return nil, provider.ErrInfoNotFound
		}
		return nil, vErr
	}
	if errors.Is(err, errMovieNotFound) {
		return nil, provider.ErrInfoNotFound
	}
	return
}

func init() {
	provider.Register(Name, New)
}
