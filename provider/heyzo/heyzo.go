package heyzo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/collection/sets"
	"github.com/metatube-community/metatube-sdk-go/common/convertor"
	"github.com/metatube-community/metatube-sdk-go/common/nextjs"
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
	baseURL      = "https://www.heyzo.com/"
	movieURL     = "https://www.heyzo.com/movie/%04s"
	movieAPIURL  = "https://www.heyzo.com/api/v1/movie?movieId=%04s"
	reviewAPIURL = "https://www.heyzo.com/api/v1/movie/%d/review?page=1&limit=%d"
	maxReviews   = 100
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
	genres := sets.NewOrderedSet[string]()
	addGenre := func(genre string) {
		if genre = strings.TrimSpace(genre); genre != "" {
			genres.Add(genre)
		}
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
			for keyword := range strings.FieldsSeq(attr.AttributeValue) {
				addGenre(keyword)
			}
		case attributeDuration:
			duration := movieDuration{}
			if json.Unmarshal([]byte(attr.AttributeValue), &duration) == nil {
				info.Runtime = parser.ParseRuntime(duration.Full)
			}
		}
	}
	info.Genres = append(info.Genres, genres.AsSlice()...)

	// Runtime (fallback)
	if info.Runtime == 0 {
		for _, file := range data.Download.Whole {
			if runtime := parser.ParseRuntime(file.Duration); runtime > 0 {
				info.Runtime = runtime
				break
			}
		}
	}

	// Preview Images (full-size only)
	images, _ := galleryImages(data)
	info.PreviewImages = append(info.PreviewImages, images...)

	// Cover+Thumb+Preview Video
	err = hzo.scrapeMoviePage(info)
	return
}

// scrapeMoviePage scrapes the cover image and the sample video, which are
// not provided by the movie API, from the video player props in the RSC
// payload of the movie page.
func (hzo *Heyzo) scrapeMoviePage(info *model.MovieInfo) (err error) {
	c := hzo.ClonedCollector()

	c.OnRequest(func(r *colly.Request) {
		// The server responds 404 Not Found for non-HTML requests.
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	})

	c.OnResponse(func(r *colly.Response) {
		var payload string
		if payload, err = nextjs.ExtractPayload(r.Body); err != nil {
			return
		}
		props := playerProps{}
		// Require the poster key, so other objects with a `src` or `type`
		// field, e.g., images, are not mistaken for the video player.
		if !nextjs.FindObject(payload, &props, "poster") {
			return
		}
		if props.Poster != "" {
			info.CoverURL = r.Request.AbsoluteURL(props.Poster)
			info.ThumbURL = info.CoverURL /* use cover as thumb */
		}
		// The player only streams HLS, there is no MP4 sample video.
		if props.Src != "" && strings.Contains(props.Type, "mpegURL") {
			info.PreviewVideoHLSURL = r.Request.AbsoluteURL(props.Src)
		}
	})

	if vErr := c.Visit(info.Homepage); vErr != nil {
		err = vErr
	}
	return
}

func (hzo *Heyzo) GetMovieReviewsByID(id string) (reviews []*model.MovieReviewDetail, err error) {
	data, err := hzo.fetchMovie(id)
	if err != nil {
		return
	}

	c := hzo.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		resp := reviewResponse{}
		if err = json.Unmarshal(r.Body, &resp); err != nil {
			return
		}
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

	// Only the first page of at most `maxReviews` (latest) reviews is fetched,
	// as fetching every review of popular movies would be too large.
	if vErr := c.Visit(fmt.Sprintf(reviewAPIURL, data.MovieSeq, maxReviews)); vErr != nil {
		err = vErr
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

// galleryImages returns the gallery images shown to guests: the full-size
// images of the first 3 lightbox slides, and the thumbnails of the locked rest.
// Ref: https://www.heyzo.com/_next/static/chunks/app/%5Blocale%5D/movie/%5BmovieId%5D/page-d1ca22c68e26fb5c.js
// JS Code:
// if(t.isBonus||t.isLimited||!f)return null;
// g=Array.from({length:21},(e,t)=>String(t+1).padStart(3,"0")),
// b=(_?g:g.slice(0,3)).map((e,l)=>({src:(0,m.bl)(t.movie_id,e,c,_)})); // lightbox slides
// g.map((e,l)=>({src:(0,m.GH)(t.movie_id,e,c,_)})) // thumbnails
func galleryImages(data *movieResponse) (images, thumbnails []string) {
	// Gallery settings of the movie page, see galleryImages.
	const (
		galleryImageCount      = 21
		guestGalleryImageCount = 3
		// isHeyzoMember is always false, as the provider browses as a guest.
		isHeyzoMember = false
	)

	if data.IsBonus || data.IsLimited || len(data.Actors) == 0 {
		return
	}
	for i := range galleryImageCount {
		n := fmt.Sprintf("%03d", i+1)
		if isHeyzoMember || i < guestGalleryImageCount {
			images = append(images, galleryImageURL(data.MovieID, n, isHeyzoMember))
		} else {
			thumbnails = append(thumbnails, galleryThumbnailURL(data.MovieID, n, isHeyzoMember))
		}
	}
	return
}

// galleryThumbnailURL (GH) and galleryImageURL (bl) port the gallery URL
// builders of the movie page.
// Ref: https://www.heyzo.com/_next/static/chunks/4067-9cce700d8142bc71.js
// JS Code:
// i=e=>"".concat(e?"member/contents":"contents","/").concat(3e3),
// s=(e,t,a,o)=>{let r=(0,n.Ib)(a);return"https://".concat(r,"/").concat(i(o),"/").concat(e,"/gallery/thumbnail_").concat(t,".jpg")},
// d=(e,t,a,o)=>{let r=(0,n.Ib)(a);return"https://".concat(r,"/").concat(i(o),"/").concat(e,"/gallery/").concat(t,".jpg")},
func galleryThumbnailURL(movieID, n string, member bool) string {
	return fmt.Sprintf("https://www.heyzo.com/%s/%s/gallery/thumbnail_%s.jpg", contentsPath(member), movieID, n)
}

func galleryImageURL(movieID, n string, member bool) string {
	return fmt.Sprintf("https://www.heyzo.com/%s/%s/gallery/%s.jpg", contentsPath(member), movieID, n)
}

func contentsPath(member bool) string {
	if member {
		return "member/contents/3000"
	}
	return "contents/3000"
}

func init() {
	provider.Register(Name, New)
}
