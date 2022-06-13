package avwiki

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/gocolly/colly/v2"

	"github.com/javtube/javtube-sdk-go/common/number"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/duga"
	"github.com/javtube/javtube-sdk-go/provider/fanza"
	"github.com/javtube/javtube-sdk-go/provider/internal/scraper"
	"github.com/javtube/javtube-sdk-go/provider/mgstage"
)

var (
	_ provider.MovieProvider = (*AVWiki)(nil)
	_ provider.MovieSearcher = (*AVWiki)(nil)
)

const (
	Name     = "AVWIKI"
	Priority = 1000
)

const (
	baseURL      = "https://www.avwiki.org/"
	movieURL     = "https://www.avwiki.org/works/%s"
	movieAPIURL  = "https://www.avwiki.org/_next/data/%s/works/%s.json?id=%s"
	searchAPIURL = "https://www.avwiki.org/_next/data/%s/works.json?q=%s"
)

type AVWiki struct {
	*scraper.Scraper
	duga    *duga.DUGA
	fanza   *fanza.FANZA
	mgstage *mgstage.MGS
}

func New() *AVWiki {
	return &AVWiki{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority,
			scraper.WithHeaders(map[string]string{
				"Referer": baseURL,
			})),
		duga:    duga.New(),
		fanza:   fanza.New(),
		mgstage: mgstage.New(),
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

	buildID, err := avw.getBuildID()
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
			for _, product := range data.PageProps.Work.Products {
				switch product.Source {
				case "fanza":
					info, err = avw.fanza.GetMovieInfoByID(product.ProductID)
				case "mgstage":
					info, err = avw.mgstage.GetMovieInfoByID(product.ProductID)
				case "duga":
					info, err = avw.duga.GetMovieInfoByID(product.ProductID)
				}
				if err != nil || info == nil || !info.Valid() {
					continue
				}
				// supplement info.
				if info.Maker == "" {
					info.Maker = product.Maker.Name
				}
				if info.Series == "" {
					info.Series = product.Series.Name
				}
				break
			}
			if info == nil || !info.Valid() {
				if err == nil {
					err = provider.ErrInfoNotFound
				}
				return
			}
			// overwrite actor info.
			if len(data.PageProps.Work.Actors) > 0 {
				var actors []string
				for _, actor := range data.PageProps.Work.Actors {
					actors = append(actors, actor.Name)
				}
				info.Actors = actors
			}
		}
	})

	err = c.Visit(fmt.Sprintf(movieAPIURL, buildID, id, url.QueryEscape(id)))
	return
}

func (avw *AVWiki) TidyKeyword(keyword string) string {
	if !number.IsUncensored(keyword) {
		return strings.ToUpper(keyword)
	}
	return ""
}

func (avw *AVWiki) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	buildID, err := avw.getBuildID()
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
				if len(work.Products) == 0 {
					// ignore if this work has no products.
					continue
				}
				results = append(results, &model.MovieSearchResult{
					ID:          work.WorkID,
					Number:      work.WorkID,
					Title:       work.Title,
					Provider:    avw.Name(),
					Homepage:    fmt.Sprintf(movieURL, work.WorkID),
					ThumbURL:    work.Products[0].ThumbnailURL,
					CoverURL:    work.Products[0].ImageURL,
					ReleaseDate: parser.ParseDate(work.MinDate),
				})
			}
		}
	})

	err = c.Visit(fmt.Sprintf(searchAPIURL, buildID, url.QueryEscape(keyword)))
	return
}

func (avw *AVWiki) getBuildID() (buildID string, err error) {
	c := avw.ClonedCollector()

	c.OnXML(`//*[@id="__NEXT_DATA__"]`, func(e *colly.XMLElement) {
		data := struct {
			BuildId string `json:"buildId"`
		}{}
		if innerErr := json.NewDecoder(strings.NewReader(e.Text)).Decode(&data); innerErr != nil {
			err = innerErr
		}
		buildID = data.BuildId
		return
	})

	err = c.Visit(baseURL)
	return
}

type Work struct {
	ID       int    `json:"id"`
	WorkID   string `json:"work_id"`
	Title    string `json:"title"`
	MinDate  string `json:"min_date"`
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
	// The stability of this provider is still unknown, so we should not register it for now.
	// provider.RegisterMovieFactory(Name, New)
}
