package kin8tengoku

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/common/rsc"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*KIN8)(nil)

const (
	Name     = "KIN8"
	Priority = 1000
)

const (
	baseURL   = "https://www.kin8tengoku.com/"
	movieURL  = "https://www.kin8tengoku.com/movie/%04s"
	reviewURL = "https://www.kin8tengoku.com/api/v1/movie/%d/review"
)

type KIN8 struct {
	*scraper.Scraper
}

func New() *KIN8 {
	return &KIN8{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (k8 *KIN8) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:kin8[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (k8 *KIN8) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return k8.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (k8 *KIN8) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	// Also accepts the legacy /moviepages/{id}/index.html URLs.
	return path.Base(strings.TrimSuffix(homepage.Path, "/index.html")), nil
}

func (k8 *KIN8) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := k8.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("KIN8-%s", id),
		Provider:      k8.Name(),
		Homepage:      rawURL,
		Maker:         "金髪天國",
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := k8.ClonedCollector()

	c.OnRequest(func(r *colly.Request) {
		// The server responds 404 Not Found for non-HTML requests.
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	})

	// Fields
	c.OnResponse(func(r *colly.Response) {
		var payload string
		if payload, err = rsc.ExtractNextPayload(r.Body); err != nil {
			return
		}
		data := movieData{}
		if rsc.FindObject(payload, &data) != nil {
			return
		}
		info.Title = data.NameUTF8
		info.CoverURL = thumbnailURL(data.MovieID, locale)
		info.ThumbURL = info.CoverURL /* use cover as thumb */
		// The movie page plays the sample video only for available movies.
		if data.AvailableDays > 0 && !data.IsPreRelease {
			info.PreviewVideoURL = sampleVideoURL(data.MovieID, locale)
		}
		info.Summary = strings.TrimSpace(data.Detail.MemoUTF8)
		info.Runtime = data.Detail.Duration / 60
		info.ReleaseDate = parser.ParseDate(data.EcpStartDate)
		if data.ActUTF8 != "" {
			info.Actors = append(info.Actors, data.ActUTF8)
		}
		hasGallery := false
		for _, category := range data.Categories {
			// Only the main categories are shown as genres.
			if category.Level == mainCategoryLevel {
				info.Genres = append(info.Genres, category.NameUTF8)
			}
			if category.KeyName == "op_flag_gallery" {
				hasGallery = true
			}
		}
		// The server renders the gallery only for movies with this category.
		if hasGallery {
			images, _ := galleryImages(data.MovieID)
			info.PreviewImages = append(info.PreviewImages, images...)
		}

		// Score
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			resp := struct {
				AvgRating float64 `json:"avgRating"`
			}{}
			if json.Unmarshal(r.Body, &resp) == nil {
				info.Score = resp.AvgRating
			}
		})
		d.Visit(fmt.Sprintf(reviewURL, data.MovieSeq))
	})

	// Cover+Thumb (fallback when the RSC payload has no movie data)
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if info.CoverURL == "" {
			info.CoverURL = e.Request.AbsoluteURL(e.Attr("content"))
			info.ThumbURL = info.CoverURL /* use cover as thumb */
		}
	})

	// Preview Video (fallback when the RSC payload has no movie data)
	c.OnXML(`//meta[@property="og:video"]`, func(e *colly.XMLElement) {
		if info.PreviewVideoURL == "" {
			info.PreviewVideoURL = e.Request.AbsoluteURL(e.Attr("content"))
		}
	})

	if vErr := c.Visit(info.Homepage); vErr != nil {
		err = vErr
	}
	return
}

// movieData is the movie object in the RSC payload of the movie page.
type movieData struct {
	MovieSeq      int    `json:"movie_seq"`
	MovieID       string `json:"movie_id"`
	ActUTF8       string `json:"act_utf8"`
	NameUTF8      string `json:"name_utf8"`
	EcpStartDate  string `json:"ecp_start_date"`
	IsPreRelease  bool   `json:"isPreRelease"`
	AvailableDays int    `json:"availableDays"`
	Categories    []struct {
		NameUTF8 string `json:"category_name_utf8"`
		KeyName  string `json:"category_key_name"`
		Level    int    `json:"LEVEL"`
	} `json:"categories"`
	Detail struct {
		MemoUTF8 string `json:"memo_utf8"`
		Duration int    `json:"duration"` // seconds
	} `json:"detail"`
}

// locale is the locale the provider browses the site in.
const locale = "ja"

// siteHost (Ib), thumbnailURL (s4) and sampleVideoURL (tQ) port the URL
// builders of the movie page, which builds the player from the movie data.
// Ref: https://www.kin8tengoku.com/_next/static/chunks/3778-8a0fcf58311a7a7e.js
// JS Code:
// let i=e=>{switch(e){case"ja":default:return"www.kin8tengoku.com";case"en":return"en.kin8tengoku.com"}}
// let i=(e,n)=>1155>Number(e)?"https://smovie.kin8tengoku.com/sample_mobile_template/".concat(e,"/hls-1800k.mp4"):"https://smovie.kin8tengoku.com/".concat(e,"/pht/").concat("ja"===n?"sample.mp4":"sample_en.mp4")
// r=(e,n)=>{let a=(0,t.Ib)(n);return"https://".concat(a,"/").concat(e,"/pht/").concat("ja"===n?"1":"main_en",".jpg")}
// t.availableDays>0&&!t.isPreRelease?(0,i.jsx)(d.A,{src:(0,L.tQ)(t.movie_id,a),poster:(0,L.s4)(t.movie_id,a)}):(0,i.jsx)(_,{thumbnailUrl:(0,L.s4)(t.movie_id,a)})
func siteHost(locale string) string {
	switch locale {
	case "en":
		return "en.kin8tengoku.com"
	default:
		return "www.kin8tengoku.com"
	}
}

func thumbnailURL(movieID, locale string) string {
	name := "main_en"
	if locale == "ja" {
		name = "1"
	}
	return fmt.Sprintf("https://%s/%s/pht/%s.jpg", siteHost(locale), movieID, name)
}

func sampleVideoURL(movieID, locale string) string {
	// Number(e) is NaN for a non-numeric ID, which fails the comparison.
	if n, err := strconv.Atoi(movieID); err == nil && n < 1155 {
		return fmt.Sprintf("https://smovie.kin8tengoku.com/sample_mobile_template/%s/hls-1800k.mp4", movieID)
	}
	name := "sample_en.mp4"
	if locale == "ja" {
		name = "sample.mp4"
	}
	return fmt.Sprintf("https://smovie.kin8tengoku.com/%s/pht/%s", movieID, name)
}

// mainCategoryLevel is the LEVEL of the categories shown on the movie page.
// Ref: https://www.kin8tengoku.com/_next/static/chunks/app/%5Blocale%5D/movie/%5BmovieId%5D/page-2df76458528bd8b0.js
// JS Code:
// c=2517,d=40, // l1:()=>d
// t.categories.map(e=>e.LEVEL!==$.l1?null:(0,i.jsx)(r(),{href:"/listpages/".concat(e.category_id,"_1"),children:"ja"===a?e.category_name_utf8:e.category_name_en}))
const mainCategoryLevel = 40

// galleryImages returns the gallery images shown to guests: the full-size
// images of the first 3 items, and the thumbnails of the locked rest.
// Ref: https://www.kin8tengoku.com/_next/static/chunks/app/%5Blocale%5D/movie/%5BmovieId%5D/page-2df76458528bd8b0.js
// JS Code:
// let p=["2","3","4","5","6","7","8","9","10","11","12","13","14","15","16"]
// A=(0,c.useMemo)(()=>Number(t)>=4075?p:p.slice(0,12),[t])
// T=(0,c.useCallback)((e,a)=>{let i="".concat(t,"/pht/").concat(e,"_lg.jpg");return k?"https://".concat(N,"/members/").concat(i):a>2?"/images/movie/gallery-only-member-".concat(m,".png"):"https://".concat(N,"/").concat(i)},[N,k,m,t]);
// A.map((e,a)=>(0,i.jsx)("button",{"data-src":T(e,a),children:(0,i.jsx)(l.default,{src:"https://".concat(N,"/").concat(t,"/pht/").concat(e,".jpg")})}))
func galleryImages(movieID string) (images, thumbnails []string) {
	// isKin8Member is always false, as the provider browses as a guest.
	const isKin8Member = false

	host := siteHost(locale)

	items := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16"}
	if n, _ := strconv.Atoi(movieID); n < 4075 {
		items = items[:12]
	}
	for i, item := range items {
		if isKin8Member {
			images = append(images, fmt.Sprintf("https://%s/members/%s/pht/%s_lg.jpg", host, movieID, item))
		} else if i <= 2 {
			images = append(images, fmt.Sprintf("https://%s/%s/pht/%s_lg.jpg", host, movieID, item))
		} else {
			thumbnails = append(thumbnails, fmt.Sprintf("https://%s/%s/pht/%s.jpg", host, movieID, item))
		}
	}
	return
}

func init() {
	provider.Register(Name, New)
}
