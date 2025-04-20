package madouqu

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*MadouQu)(nil)
	_ provider.MovieSearcher = (*MadouQu)(nil)
)

const (
	Name     = "MadouQu"
	Priority = 0 // Disabled by default, use `export MT_MOVIE_PROVIDER_MADOUQU__PRIORITY=1000` to enable.
)

const (
	baseURL   = "https://madouqu.com/"
	movieURL  = "https://madouqu.com/%s/"
	searchURL = "https://madouqu.com/?s=%s"
)

type MadouQu struct {
	*scraper.Scraper
}

func New() *MadouQu {
	return &MadouQu{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Chinese)}
}

func (mdq *MadouQu) SetRequestTimeout(_ time.Duration) {
	mdq.Scraper.SetRequestTimeout(10 * time.Second) // force timeout setting.
}

func (mdq *MadouQu) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return mdq.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (mdq *MadouQu) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (mdq *MadouQu) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := mdq.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Provider:      mdq.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := mdq.ClonedCollector()

	c.OnXML(`//article[starts-with(@id,'post')]//div[@class="container"]//p`, func(e *colly.XMLElement) {
		if src := e.ChildAttr(`./img`, "src"); src != "" {
			info.CoverURL = ExtractImgSrc(src)
			return
		}

		switch {
		case strings.Contains(e.Text, "番號"):
			_, nb, _ := strings.Cut(e.Text, "：")
			info.Number = nb
		case strings.Contains(e.Text, "片名"):
			_, title, _ := strings.Cut(e.Text, "：")
			info.Title = title
		case strings.Contains(e.Text, "女郎"):
			_, actors, _ := strings.Cut(e.Text, "：")
			for _, actor := range strings.Split(actors, "、") {
				info.Actors = append(info.Actors, strings.TrimSpace(actor))
			}
		}
	})

	actorTags := make([]string, 0, 2)

	// Tags (Actors)
	c.OnXML(`//article[starts-with(@id,'post')]//div[@class="entry-tags"]/a`, func(e *colly.XMLElement) {
		actorTags = append(actorTags, strings.TrimSpace(e.Text))
	})

	// Maker
	c.OnXML(`//article[starts-with(@id,'post')]//span[@class="meta-category"]`, func(e *colly.XMLElement) {
		if info.Maker == "" {
			info.Maker = strings.TrimSpace(e.ChildText(`./a[1]`))
		}
	})

	// Fallback
	c.OnScraped(func(_ *colly.Response) {
		// Number = Upper ID
		if info.Number == "" {
			info.Number = parser.ParseIDToNumber(info.ID)
		}

		// Thumb Image
		if info.ThumbURL == "" {
			info.ThumbURL = info.CoverURL // same as cover
		}

		// Actors
		if len(info.Actors) == 0 && len(actorTags) > 0 {
			info.Actors = actorTags // fallback
		}
	})

	err = c.Visit(info.Homepage)
	return
}

func (mdq *MadouQu) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToUpper(keyword)
}

func (mdq *MadouQu) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := mdq.ClonedCollector()

	c.OnXML(`//article[starts-with(@id, 'post')]`, func(e *colly.XMLElement) {
		link := e.ChildAttr(`.//h2/a`, "href")
		id, idErr := mdq.ParseMovieIDFromURL(link)
		if idErr != nil {
			return
		}

		origTitle := e.ChildAttr(`.//h2/a`, "title")
		nb, title, _ := strings.Cut(origTitle, " ")

		if !regexp.MustCompile(`^(?i)[A-Z0-9_-]+$`).MatchString(nb) {
			nb = parser.ParseIDToNumber(id)
			title = origTitle
		}

		thumb := ExtractImgSrc(e.ChildAttr(`.//div[@class="entry-media"]//a/img`, "data-src"))

		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Number:      nb,
			Title:       title,
			Provider:    mdq.Name(),
			Homepage:    link,
			ThumbURL:    thumb,
			CoverURL:    thumb, // same as thumb
			ReleaseDate: parser.ParseDate(e.ChildAttr(`.//li/time`, "datetime")),
		})
	})

	if vErr := c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword))); vErr != nil {
		err = vErr
	}
	return
}

func ExtractImgSrc(src string) string {
	u, err := url.Parse(src)
	if err != nil {
		return src
	}
	if ss := regexp.MustCompile(`(https?://.+$)`).FindStringSubmatch(u.Path); len(ss) > 0 {
		return ss[1]
	}
	return src
}

func init() {
	provider.Register(Name, New)
}
