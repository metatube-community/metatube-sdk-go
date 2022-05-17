package d2pass

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/fetch"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
)

// API Paths
const (
	moviePath              = "/movies/%s/"
	movieDetailPath        = "/dyn/phpauto/movie_details/movie_id/%s.json"
	movieGalleryPath       = "/dyn/dla/json/movie_gallery/%s.json"
	movieLegacyGalleryPath = "/dyn/phpauto/movie_galleries/movie_id/%s.json"
)

type Core struct {
	*provider.Scraper

	// BaseURL is the base URL of provider website.
	BaseURL string

	// Default values
	DefaultPriority int
	DefaultName     string
	DefaultMaker    string

	// Gallery paths
	GalleryPath       string
	LegacyGalleryPath string
}

func (core *Core) Init() {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.IgnoreRobotsTxt(),
		colly.UserAgent(random.UserAgent()),
		colly.Headers(map[string]string{
			"Content-Type": "application/json",
		}))
	c.SetCookies(core.BaseURL, []*http.Cookie{
		{Name: "ageCheck", Value: "1"},
	})
	core.Scraper = provider.NewScraper(core.DefaultName, core.DefaultPriority, c)
}

func (core *Core) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return core.GetMovieInfoByURL(fetch.JoinURL(core.BaseURL, fmt.Sprintf(moviePath, id)))
}

func (core *Core) GetMovieInfoByURL(u string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	id := path.Base(homepage.Path)

	info = &model.MovieInfo{
		Provider:      core.Name(),
		Homepage:      homepage.String(),
		Maker:         core.DefaultMaker,
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := core.Collector()

	c.OnResponse(func(r *colly.Response) {
		data := struct {
			ActressesJa []string
			AvgRating   float64
			Desc        string
			Duration    int
			Gallery     bool
			HasGallery  bool
			MovieID     string
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
				info.Tags = data.UCNAME
			}
			if len(data.ActressesJa) > 0 {
				info.Actors = data.ActressesJa
			}
			if len(data.SampleFiles) > 0 {
				sort.SliceStable(data.SampleFiles, func(i, j int) bool {
					return data.SampleFiles[i].FileSize < data.SampleFiles[j].FileSize
				})
				info.PreviewVideoURL = r.Request.AbsoluteURL(data.SampleFiles[len(data.SampleFiles)-1].URL)
			}
			for _, thumb := range []string{
				data.ThumbUltra, data.ThumbHigh,
				data.ThumbMed, data.ThumbLow,
			} {
				if thumb != "" {
					info.CoverURL = r.Request.AbsoluteURL(thumb)
					info.ThumbURL = info.CoverURL /* use thumb as cover */
					break
				}
			}
			// Gallery Code:
			// Ref: https://www.1pondo.tv/js/movieDetail.0155a1b9.js:formatted#1452
			//
			// return Object.prototype.hasOwnProperty.call(this.movieDetail, "Gallery") && this.movieDetail.Gallery ?
			// this.hasGallery = !0 : this.movieDetail.HasGallery && (this.hasGallery = !0, this.legacyGallery = !0),
			// e.getMovieGallery(this.movieDetail.MovieID, this.legacyGallery);
			// Preview Images
			if data.Gallery {
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
			} else if data.HasGallery /* Legacy Gallery */ {
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
						//movieGallery: {
						//   sampleURLs: {
						//	   preview: "/assets/sample/{MOVIE_ID}/thum_106/{FILENAME}.jpg",
						//	   fullsize: "/assets/sample/{MOVIE_ID}/popu/{FILENAME}.jpg",
						//	   movieIdKey: "MovieID"
						//   }
						//}
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

	err = c.Visit(fetch.JoinURL(info.Homepage, fmt.Sprintf(movieDetailPath, id)))
	return
}
