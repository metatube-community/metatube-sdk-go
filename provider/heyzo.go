package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube/model"
	"github.com/javtube/javtube/util"
)

var _ Provider = (*Heyzo)(nil)

type Heyzo struct {
	BaseURL  string
	MovieURL string
}

func NewHeyzo() Provider {
	return &Heyzo{
		BaseURL:  "https://www.heyzo.com/",
		MovieURL: "https://www.heyzo.com/moviepages/%04s/index.html",
	}
}

func (hzo *Heyzo) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return hzo.GetMovieInfoByLink(fmt.Sprintf(hzo.MovieURL, id))
}

func (hzo *Heyzo) GetMovieInfoByLink(link string) (info *model.MovieInfo, err error) {
	homepage, err := url.Parse(strings.TrimRight(link, "/"))
	if err != nil {
		return nil, err
	}
	id := path.Base(path.Dir(homepage.Path))

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("HEYZO-%s", id),
		Homepage:      homepage.String(),
		Actors:        []string{},
		PreviewImages: []string{},
		Tags:          []string{},
	}

	c := colly.NewCollector(colly.UserAgent(UA))

	// JSON
	c.OnXML(`//script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			Name          string `json:"name"`
			Image         string `json:"image"`
			Description   string `json:"description"`
			ReleasedEvent struct {
				StartDate string `json:"startDate"`
			} `json:"releasedEvent"`
			Video struct {
				Duration string `json:"duration"`
				Actor    string `json:"actor"`
				Provider string `json:"provider"`
			} `json:"video"`
			AggregateRating struct {
				RatingValue string `json:"ratingValue"`
			} `json:"aggregateRating"`
		}{}
		if json.Unmarshal([]byte(e.Text), &data) == nil {
			info.Title = data.Name
			info.Summary = data.Description
			info.CoverURL = e.Request.AbsoluteURL(data.Image)
			info.ThumbURL = info.CoverURL /* use cover as thumb */
			info.Publisher = data.Video.Provider
			info.ReleaseDate = util.ParseDate(data.ReleasedEvent.StartDate)
			info.Duration = util.ParseDuration(strings.TrimPrefix(data.Video.Duration, "PT"))
			info.Score = util.ParseScore(data.AggregateRating.RatingValue)
			if data.Video.Actor != "" {
				info.Actors = []string{data.Video.Actor}
			}
		}
	})

	// Title
	c.OnXML(`//*[@id="movie"]/h1`, func(e *colly.XMLElement) {
		if info.Title == "" {
			info.Title = strings.Fields(e.Text)[0]
		}
	})

	// Summary
	c.OnXML(`//p[@class="memo"]`, func(e *colly.XMLElement) {
		if info.Summary == "" {
			info.Summary = strings.TrimSpace(e.Text)
		}
	})

	// Fields
	c.OnXML(`//table[@class="movieInfo"]/tbody/tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "公開日":
			info.ReleaseDate = util.ParseDate(e.ChildText(`.//td[2]`))
		case "出演":
			info.Actors = e.ChildTexts(`.//td[2]/a/span`)
		case "シリーズ":
			info.Series = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "評価":
			info.Score = util.ParseScore(e.ChildText(`.//span[@itemprop="ratingValue"]`))
		}
	})

	// Tags
	c.OnXML(`//ul[@class="tag-keyword-list"]`, func(e *colly.XMLElement) {
		info.Tags = e.ChildTexts(`.//li/a`)
	})

	// Video+Duration
	c.OnXML(`//script[@type="text/javascript"]`, func(e *colly.XMLElement) {
		if info.PreviewVideoURL != "" && info.Duration != 0 {
			return
		}
		// Sample Video
		if strings.Contains(e.Text, "emvideo") {
			if sub := regexp.MustCompile(`emvideo = "(.+?)";`).FindStringSubmatch(e.Text); len(sub) == 2 {
				info.PreviewVideoURL = e.Request.AbsoluteURL(strings.ReplaceAll(sub[1], "_low", ""))
			}
		}
		// Duration
		if strings.Contains(e.Text, "o = {") {
			if sub := regexp.MustCompile(`o = (\{.+?});`).FindStringSubmatch(e.Text); len(sub) == 2 {
				data := struct {
					Full string `json:"full"`
				}{}
				if json.Unmarshal([]byte(sub[1]), &data) == nil {
					info.Duration = util.ParseDuration(
						regexp.MustCompile(`(\d\d):(\d\d):(\d\d)`).
							ReplaceAllString(data.Full, "${1}h${2}m${3}s"))
				}
			}
		}
	})

	// Preview Images
	c.OnXML(`//div[@class="sample-images yoxview"]/script`, func(e *colly.XMLElement) {
		for _, sub := range regexp.MustCompile(`"(/contents/.+/\d+?\.\w+?)"`).FindAllStringSubmatch(e.Text, -1) {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(sub[1]))
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (hzo *Heyzo) SearchMovie(keyword string) (results []*model.SearchResult, err error) {
	return nil, errors.New("no search support for Heyzo")
}
