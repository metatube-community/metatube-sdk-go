package caribbeancompr

import (
	"fmt"
	"regexp"

	"github.com/gocolly/colly/v2"
	"github.com/javtube/javtube-sdk-go/common/random"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
	"github.com/javtube/javtube-sdk-go/provider/caribbeancom"
)

const (
	Name     = "CARIBBEANCOMPR"
	Priority = 1000
)

const (
	baseURL  = "https://www.caribbeancompr.com/"
	movieURL = "https://www.caribbeancompr.com/moviepages/%s/index.html"
)

type CaribbeancomPR struct {
	*caribbeancom.Caribbeancom
}

func New() *CaribbeancomPR {
	return &CaribbeancomPR{
		Caribbeancom: &caribbeancom.Caribbeancom{
			Scraper: provider.NewScraper(Name, Priority, colly.NewCollector(
				colly.AllowURLRevisit(),
				colly.IgnoreRobotsTxt(),
				colly.DetectCharset(),
				colly.UserAgent(random.UserAgent()))),
		},
	}
}

func (carib *CaribbeancomPR) NormalizeID(id string) string {
	if regexp.MustCompile(`^\d{6}_\d{3}$`).MatchString(id) {
		return id
	}
	return ""
}

func (carib *CaribbeancomPR) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return carib.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func init() {
	provider.RegisterMovieFactory(Name, New)
}
