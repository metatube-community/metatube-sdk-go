package avbase

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
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
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
	_ provider.MovieProvider = (*AVBase)(nil)
	_ provider.MovieSearcher = (*AVBase)(nil)
	_ provider.Fetcher       = (*AVBase)(nil)
)

const (
	Name     = "AVBASE"
	Priority = 1000 - 4
)

const (
	baseURL      = "https://www.avbase.net/"
	movieURL     = "https://www.avbase.net/works/%s"
	movieAPIURL  = "https://www.avbase.net/_next/data/%s/works/%s.json?id=%s"
	searchAPIURL = "https://www.avbase.net/_next/data/%s/works.json?q=%s"
)

type AVBase struct {
	*fetch.Fetcher
	*scraper.Scraper
	single    *singledo.Single
	providers map[string]provider.MovieProvider
}

func New() *AVBase {
	return &AVBase{
		Fetcher: fetch.Default(&fetch.Config{SkipVerify: true}),
		Scraper: scraper.NewDefaultScraper(
			Name, baseURL, Priority, language.Japanese,
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

func (ab *AVBase) NormalizeMovieID(id string) string {
	if !strings.Contains(id, ":") {
		return strings.ToUpper(id)
	}
	ss := strings.SplitN(id, ":", 2)
	prefix, workID := ss[0], ss[1]
	return ab.JoinPrefixID(prefix, workID)
}

func (ab *AVBase) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return ab.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (ab *AVBase) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return ab.NormalizeMovieID(path.Base(homepage.Path)), nil
}

func (ab *AVBase) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := ab.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	buildID, err := ab.GetBuildID()
	if err != nil {
		return
	}

	c := ab.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		data := struct {
			PageProps struct {
				Work workResponse `json:"work"`
			} `json:"pageProps"`
		}{}
		if err = json.Unmarshal(r.Body, &data); err == nil {
			workInfo, _ := ab.getMovieInfoFromWork(data.PageProps.Work)
			srcInfo, srcErr := ab.getMovieInfoFromSource(data.PageProps.Work)
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
			if info.Summary == "" {
				info.Summary = workInfo.Summary
			}
			if len(info.Genres) == 0 {
				info.Genres = workInfo.Genres
			}
			// replace actor names.
			if len(workInfo.Actors) > 0 {
				info.Actors = workInfo.Actors
			}
			// prefer workID number.
			if workInfo.Number != "" {
				info.Number = workInfo.Number
			}
			// choose right ID for info.
			if len(workInfo.ID) > len(id) && strings.Contains(workInfo.ID, ":") {
				id = workInfo.ID
			}
		}
	})

	c.OnScraped(func(_ *colly.Response) {
		if info != nil {
			// As a provider wrapper.
			info.ID = id
			info.Provider = ab.Name()
			info.Homepage = rawURL
		}
	})

	if vErr := c.Visit(fmt.Sprintf(movieAPIURL, buildID, id, url.QueryEscape(id))); vErr != nil {
		err = vErr
	}
	return
}

func (ab *AVBase) getMovieInfoFromWork(work workResponse) (info *model.MovieInfo, err error) {
	info = &model.MovieInfo{
		ID:            ab.JoinPrefixID(work.Prefix, work.WorkID),
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
		if info.Summary == "" {
			info.Summary = product.ItemInfo.Description
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
	for _, cast := range work.Casts {
		info.Actors = append(info.Actors, cast.Actor.Name)
	}
	return
}

func (ab *AVBase) getMovieInfoFromSource(work workResponse) (info *model.MovieInfo, err error) {
	for _, product := range work.Products {
		movieProvider, ok := ab.providers[product.Source]
		if !ok {
			continue
		}
		info, err = movieProvider.GetMovieInfoByID(product.ProductID)
		if err != nil || info == nil || !info.IsValid() {
			continue
		}
		break
	}
	if info == nil || !info.IsValid() {
		if err == nil {
			err = provider.ErrInfoNotFound
		}
	}
	return
}

func (ab *AVBase) NormalizeMovieKeyword(keyword string) string {
	if number.IsUncensored(keyword) || number.IsFC2(keyword) {
		return "" // no uncensored support.
	}
	return strings.ToUpper(keyword)
}

func (ab *AVBase) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	buildID, err := ab.GetBuildID()
	if err != nil {
		return
	}

	c := ab.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		data := struct {
			PageProps struct {
				Works []workResponse `json:"works"`
			} `json:"pageProps"`
		}{}
		if json.Unmarshal(r.Body, &data) == nil {
			for _, work := range data.PageProps.Works {
				sort.SliceStable(work.Products, func(i, j int) bool {
					return work.Products[i].Source > work.Products[j].Source
				})
				index := -1
				for i, product := range work.Products {
					if _, ok := ab.providers[product.Source]; ok {
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
					ID:          ab.JoinPrefixID(work.Prefix, work.WorkID),
					Number:      work.WorkID,
					Title:       work.Title,
					Provider:    ab.Name(),
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

func (ab *AVBase) JoinPrefixID(prefix, workID string) string {
	if strings.TrimSpace(prefix) == "" {
		return workID
	}
	return fmt.Sprintf("%s:%s", prefix, workID)
}

func (ab *AVBase) GetBuildID() (string, error) {
	v, err, _ := ab.single.Do(func() (any, error) {
		return ab.getBuildID()
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (ab *AVBase) getBuildID() (buildID string, err error) {
	defer func() {
		if err == nil && buildID == "" {
			err = errors.New("empty build id")
		}
	}()

	c := ab.ClonedCollector()

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

func init() {
	// The stability of this provider is still unknown.
	provider.Register(Name, New)
}
