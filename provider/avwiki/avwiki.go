package avwiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/common/singledo"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/duga"
	"github.com/metatube-community/metatube-sdk-go/provider/fanza"
	"github.com/metatube-community/metatube-sdk-go/provider/getchu"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
	"github.com/metatube-community/metatube-sdk-go/provider/mgstage"
	"github.com/metatube-community/metatube-sdk-go/provider/pcolle"
)

var (
	_ provider.MovieProvider = (*AVWiki)(nil)
	_ provider.MovieSearcher = (*AVWiki)(nil)
)

const (
	Name     = "AVWIKI"
	Priority = 1000 - 4
)

const (
	baseURL      = "https://www.avwiki.org/"
	movieURL     = "https://www.avwiki.org/works/%s"
	movieAPIURL  = "https://www.avwiki.org/_next/data/%s/works/%s.json?id=%s"
	searchAPIURL = "https://www.avwiki.org/_next/data/%s/works.json?q=%s"
)

type AVWiki struct {
	*scraper.Scraper
	single    *singledo.Single
	providers map[string]provider.MovieProvider
}

func New() *AVWiki {
	return &AVWiki{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority,
			scraper.WithHeaders(map[string]string{
				"Referer": baseURL,
			})),
		single: singledo.NewSingle(2 * time.Hour),
		providers: map[string]provider.MovieProvider{
			"duga":    duga.New(),
			"fanza":   fanza.New(),
			"getchu":  getchu.New(),
			"mgstage": mgstage.New(),
			"pcolle":  pcolle.New(),
		},
	}
}

func (avw *AVWiki) NormalizeID(id string) string { return strings.ToUpper(id) }

func (avw *AVWiki) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return avw.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (avw *AVWiki) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return avw.NormalizeID(path.Base(homepage.Path)), nil
}

func (avw *AVWiki) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := avw.ParseIDFromURL(rawURL)
	if err != nil {
		return
	}

	buildID, err := avw.GetBuildID()
	if err != nil {
		return
	}

	c := avw.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		data := struct {
			PageProps struct {
				Work Work `json:"work"`
			} `json:"pageProps"`
		}{}
		if err = json.Unmarshal(r.Body, &data); err == nil {
			workInfo, _ := avw.getMovieInfoFromWork(data.PageProps.Work)
			srcInfo, srcErr := avw.getMovieInfoFromSource(data.PageProps.Work)
			if srcErr != nil {
				info = workInfo /* ignore error and fallback to work info */
				return
			}
			// use source info.
			info = srcInfo
			// supplement info fields.
			if info.Maker == "" {
				info.Maker = workInfo.Maker
			}
			if info.Label == "" {
				info.Label = workInfo.Label
			}
			if info.Series == "" {
				info.Series = workInfo.Series
			}
			if len(info.Genres) == 0 {
				info.Genres = workInfo.Genres
			}
			// replace actor names.
			if len(workInfo.Actors) > 0 {
				info.Actors = workInfo.Actors
			}
		}
	})

	c.OnScraped(func(_ *colly.Response) {
		if info != nil {
			// As a provider wrapper.
			info.ID = id
			info.Provider = avw.Name()
			info.Homepage = rawURL
		}
	})

	if vErr := c.Visit(fmt.Sprintf(movieAPIURL, buildID, id, url.QueryEscape(id))); vErr != nil {
		err = vErr
	}
	return
}

func (avw *AVWiki) getMovieInfoFromWork(work Work) (info *model.MovieInfo, err error) {
	info = &model.MovieInfo{
		Number:        work.WorkID,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}
	sort.SliceStable(work.Products, func(i, j int) bool {
		// we want mgs > fanza > duga, etc.
		return work.Products[i].Source > work.Products[j].Source
	})
	for _, product := range work.Products {
		if info.Title == "" {
			info.Title = product.Title
		}
		if info.CoverURL == "" {
			info.CoverURL = product.ImageURL
		}
		if info.ThumbURL == "" {
			info.ThumbURL = product.ThumbnailURL
		}
		if info.Maker == "" {
			info.Maker = product.Maker.Name
		}
		if info.Label == "" {
			info.Label = product.Label.Name
		}
		if info.Series == "" {
			info.Series = product.Series.Name
		}
		if time.Time(info.ReleaseDate).IsZero() {
			info.ReleaseDate = parser.ParseDate(product.Date)
		}
		if len(info.PreviewImages) == 0 {
			for _, sample := range product.SampleImageURLS {
				if sample.L == "" {
					continue
				}
				info.PreviewImages = append(info.PreviewImages, sample.L)
			}
		}
	}
	for _, genre := range work.Genres {
		info.Genres = append(info.Genres, genre.Name)
	}
	for _, actor := range work.Actors {
		info.Actors = append(info.Actors, actor.Name)
	}
	return
}

func (avw *AVWiki) getMovieInfoFromSource(work Work) (info *model.MovieInfo, err error) {
	for _, product := range work.Products {
		movieProvider, ok := avw.providers[product.Source]
		if !ok {
			continue
		}
		info, err = movieProvider.GetMovieInfoByID(product.ProductID)
		if err != nil || info == nil || !info.Valid() {
			continue
		}
		break
	}
	if info == nil || !info.Valid() {
		if err == nil {
			err = provider.ErrInfoNotFound
		}
	}
	return
}

func (avw *AVWiki) NormalizeKeyword(keyword string) string {
	if number.IsUncensored(keyword) || number.IsFC2(keyword) {
		return "" // no uncensored support.
	}
	return strings.ToUpper(keyword)
}

func (avw *AVWiki) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	buildID, err := avw.GetBuildID()
	if err != nil {
		return
	}

	c := avw.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		data := struct {
			PageProps struct {
				Works []Work `json:"works"`
			} `json:"pageProps"`
		}{}
		if json.Unmarshal(r.Body, &data) == nil {
			for _, work := range data.PageProps.Works {
				sort.SliceStable(work.Products, func(i, j int) bool {
					return work.Products[i].Source > work.Products[j].Source
				})
				index := -1
				for i, product := range work.Products {
					if _, ok := avw.providers[product.Source]; ok {
						index = i
						break
					}
				}
				if index < 0 {
					// ignore if this work has no products or
					// no suitable source providers.
					continue
				}
				result := &model.MovieSearchResult{
					ID:          work.WorkID,
					Number:      work.WorkID,
					Title:       work.Title,
					Provider:    avw.Name(),
					Homepage:    fmt.Sprintf(movieURL, work.WorkID),
					ThumbURL:    work.Products[index].ThumbnailURL,
					CoverURL:    work.Products[index].ImageURL,
					ReleaseDate: parser.ParseDate(work.MinDate),
				}
				for _, actor := range work.Actors {
					result.Actors = append(result.Actors, actor.Name)
				}
				results = append(results, result)
			}
		}
	})

	err = c.Visit(fmt.Sprintf(searchAPIURL, buildID, url.QueryEscape(keyword)))
	return
}

func (avw *AVWiki) GetBuildID() (string, error) {
	v, err, _ := avw.single.Do(func() (any, error) {
		return avw.getBuildID()
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (avw *AVWiki) getBuildID() (buildID string, err error) {
	defer func() {
		if err == nil && buildID == "" {
			err = errors.New("empty build id")
		}
	}()

	c := avw.ClonedCollector()

	c.OnXML(`//*[@id="__NEXT_DATA__"]`, func(e *colly.XMLElement) {
		data := struct {
			BuildId string `json:"buildId"`
		}{}
		if err = json.NewDecoder(strings.NewReader(e.Text)).Decode(&data); err == nil {
			buildID = data.BuildId
		}
	})

	if vErr := c.Visit(baseURL); vErr != nil {
		err = vErr
	}
	return
}

type Work struct {
	ID      int    `json:"id"`
	WorkID  string `json:"work_id"`
	Title   string `json:"title"`
	MinDate string `json:"min_date"`
	Genres  []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Products []struct {
		ID           int    `json:"id"`
		ProductID    string `json:"product_id"`
		URL          string `json:"url"`
		Title        string `json:"title"`
		Source       string `json:"source"`
		ImageURL     string `json:"image_url"`
		ThumbnailURL string `json:"thumbnail_url"`
		Date         string `json:"date"`
		Maker        struct {
			Name string `json:"name"`
		} `json:"maker"`
		Label struct {
			Name string `json:"name"`
		} `json:"label"`
		Series struct {
			Name string `json:"name"`
		} `json:"series"`
		SampleImageURLS []struct {
			S string `json:"s"`
			L string `json:"l"`
		} `json:"sample_image_urls"`
	} `json:"products"`
	Actors []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		ImageURL string `json:"image_url"`
	} `json:"actors"`
}

func init() {
	// The stability of this provider is still unknown.
	provider.RegisterMovieFactory(Name, New)
}
