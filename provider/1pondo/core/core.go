package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/nlnwa/whatwg-url/url"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

// API Paths
const (
	movieDetailPath        = "/dyn/phpauto/movie_details/movie_id/%s.json"
	movieGalleryPath       = "/dyn/dla/json/movie_gallery/%s.json"
	movieLegacyGalleryPath = "/dyn/phpauto/movie_galleries/movie_id/%s.json"
)

type Core struct {
	*scraper.Scraper

	// URLs
	BaseURL        string
	MovieURL       string
	SampleVideoURL string

	// Values
	DefaultPriority int
	DefaultName     string
	DefaultMaker    string

	// Paths
	GalleryPath       string
	LegacyGalleryPath string
}

func (core *Core) Init() *Core {
	core.Scraper = scraper.NewDefaultScraper(core.DefaultName, core.BaseURL, core.DefaultPriority,
		scraper.WithHeaders(map[string]string{
			"Content-Type": "application/json",
		}),
		scraper.WithCookies(core.BaseURL, []*http.Cookie{
			{Name: "ageCheck", Value: "1"},
		}),
	)
	return core
}

func (core *Core) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return core.GetMovieInfoByURL(fmt.Sprintf(core.MovieURL, id))
}

func (core *Core) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := urlParser.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Pathname()), nil
}

func (core *Core) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := core.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		Provider:      core.Name(),
		Homepage:      rawURL,
		Maker:         core.DefaultMaker,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := core.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		data := struct {
			ActressesJa []string
			AvgRating   float64
			Desc        string
			Duration    int
			Gallery     bool
			HasGallery  bool
			MovieID     string
			MovieThumb  string
			Release     string
			Series      string
			ThumbHigh   string
			ThumbLow    string
			ThumbMed    string
			ThumbUltra  string
			Title       string
			UCNAME      []string
			SampleFiles []struct {
				FileSize int
				URL      string
			}
		}{}
		if err = json.Unmarshal(r.Body, &data); err == nil {
			info.ID = data.MovieID
			info.Number = info.ID
			info.Title = data.Title
			info.Summary = data.Desc
			info.Series = data.Series
			info.ReleaseDate = parser.ParseDate(data.Release)
			info.Runtime = int((time.Duration(data.Duration) * time.Second).Minutes())
			if data.AvgRating <= 5 {
				info.Score = data.AvgRating
			}
			if len(data.UCNAME) > 0 {
				info.Genres = data.UCNAME
			}
			if len(data.SampleFiles) > 0 {
				// hasCensoredSample: function() {
				//   var t = Date.parse(this.movie.Release) / 1e3;
				//   return t >= 1489622400
				// }
				sort.SliceStable(data.SampleFiles, func(i, j int) bool {
					return data.SampleFiles[i].FileSize < data.SampleFiles[j].FileSize
				})
				info.PreviewVideoURL = r.Request.AbsoluteURL(data.SampleFiles[len(data.SampleFiles)-1].URL)
				// "hls-default": {
				//   url: "https://fms.1pondo.tv/sample/{MOVIE_ID}/mb.m3u8",
				//   deliveryType: "hls",
				//   mimeType: "application/vnd.apple.mpegurl",
				//   movieIdKey: "MovieID"
				// },
				info.PreviewVideoHLSURL = fmt.Sprintf(core.SampleVideoURL, data.MovieID)
			}
			for _, actor := range data.ActressesJa {
				if actor := strings.Trim(actor, "-"); actor != "" {
					info.Actors = append(info.Actors, actor)
				}
			}
			for _, thumb := range []string{
				data.ThumbUltra, data.ThumbHigh,
				data.ThumbMed, data.ThumbLow,
			} {
				if thumb != "" {
					if re := regexp.MustCompile(`^https?:///`); re.MatchString(thumb) {
						// Fix a rare case that causes incomplete url issue.
						// e.g.: "https:///moviepages/071319_870/images/str.jpg"
						thumb = re.ReplaceAllString(thumb, "/")
					}
					info.CoverURL = r.Request.AbsoluteURL(thumb)
					info.ThumbURL = info.CoverURL /* use thumb as cover */
					break
				}
			}
			if data.MovieThumb != "" {
				info.ThumbURL = r.Request.AbsoluteURL(data.MovieThumb)
			}
			// Gallery Code:
			// Ref: https://www.1pondo.tv/js/movieDetail.0155a1b9.js:formatted#1452
			//
			// return Object.prototype.hasOwnProperty.call(this.movieDetail, "Gallery") && this.movieDetail.Gallery ?
			// this.hasGallery = !0 : this.movieDetail.HasGallery && (this.hasGallery = !0, this.legacyGallery = !0),
			// e.getMovieGallery(this.movieDetail.MovieID, this.legacyGallery);
			// Preview Images
			if data.Gallery && core.GalleryPath != "" {
				d := c.Clone()
				d.OnResponse(func(r *colly.Response) {
					galleries := struct {
						Rows []struct {
							Img       string
							Protected bool
						}
					}{}
					if json.Unmarshal(r.Body, &galleries) == nil {
						//for (var c = 0; c < this.gallery.Rows.length; c += 1) {
						//   this.$set(this.gallery.Rows[c], "idx", c);
						//   var u = !1;
						//   (!t || t && i) && (u = !0);
						//   var d = u || !this.gallery.Rows[c].Protected;
						//   this.$set(this.gallery.Rows[c], "canViewFull", d);
						//   var v = "/dyn/dla/images/".concat(this.gallery.Rows[c].Img)
						//	 , f = "".concat(a, "/dyn/dla/images/").concat(this.gallery.Rows[c].Img);
						//   this.$set(this.gallery.Rows[c], "PreviewURL", v.replace("member", "sample").replace(/\.jpg/, "__@120.jpg")),
						//   this.$set(this.gallery.Rows[c], "FullsizeURL", f),
						//   this.gallery.Rows[c].Protected && d && i && this.$set(this.gallery.Rows[c], "FullsizeURL", f += "?m=".concat(i))
						//}
						for _, row := range galleries.Rows {
							if !row.Protected {
								info.PreviewImages = append(info.PreviewImages,
									r.Request.AbsoluteURL(fmt.Sprintf(core.GalleryPath, row.Img)))
							}
						}
					}
				})
				d.Visit(r.Request.AbsoluteURL(fmt.Sprintf(movieGalleryPath, id)))
			} else if data.HasGallery /* Legacy Gallery */ && core.LegacyGalleryPath != "" {
				d := c.Clone()
				d.OnResponse(func(r *colly.Response) {
					galleries := struct {
						Rows []struct {
							MovieID   string
							Filename  string
							Protected bool
						}
					}{}
					if json.Unmarshal(r.Body, &galleries) == nil {
						for _, row := range galleries.Rows {
							if !row.Protected {
								info.PreviewImages = append(info.PreviewImages,
									r.Request.AbsoluteURL(fmt.Sprintf(core.LegacyGalleryPath,
										row.MovieID, row.Filename)))
							}
						}
					}
				})
				d.Visit(r.Request.AbsoluteURL(fmt.Sprintf(movieLegacyGalleryPath, id)))
			}
		}
	})

	if vErr := c.Visit(urlJoin(info.Homepage, fmt.Sprintf(movieDetailPath, id))); vErr != nil {
		err = vErr
	}
	return
}

var urlParser = url.NewParser(url.WithPercentEncodeSinglePercentSign())

func urlJoin(url, path string) string {
	absURL, err := urlParser.ParseRef(url, path)
	if err != nil {
		return ""
	}
	return absURL.Href(false)
}
