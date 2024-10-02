// Deprecated: X/sList is deprecated due to its outdated data and WAF.
package xslist

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	dt "gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.ActorProvider = (*XsList)(nil)
	_ provider.ActorSearcher = (*XsList)(nil)
)

const (
	Name     = "XsList" // `X/sList`
	Priority = 1000
)

const (
	baseURL   = "https://xslist.org/"
	actorURL  = "https://xslist.org/zh/model/%s.html"
	searchURL = "https://xslist.org/search?query=%s&lg=zh"
)

type XsList struct {
	*scraper.Scraper
}

func New() *XsList {
	return &XsList{scraper.NewDefaultScraper(Name, baseURL, Priority, scraper.WithDisableCookies())}
}

func (xsl *XsList) GetActorInfoByID(id string) (info *model.ActorInfo, err error) {
	return xsl.GetActorInfoByURL(fmt.Sprintf(actorURL, id))
}

func (xsl *XsList) ParseActorIDFromURL(rawURL string) (id string, err error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if ext := path.Ext(homepage.Path); ext != "" {
		id = path.Base(homepage.Path[:len(homepage.Path)-len(ext)])
	}
	return
}

func (xsl *XsList) GetActorInfoByURL(rawURL string) (info *model.ActorInfo, err error) {
	id, err := xsl.ParseActorIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.ActorInfo{
		ID:       id,
		Provider: xsl.Name(),
		Homepage: rawURL,
		Aliases:  []string{},
		Images:   []string{},
	}

	c := xsl.ClonedCollector()

	// Name
	c.OnXML(`//*[@id="sss1"]/header/h1/span`, func(e *colly.XMLElement) {
		info.Name = e.Text
	})

	// Aliases
	c.OnXML(`//*[@id="sss1"]/p/span`, func(e *colly.XMLElement) {
		info.Aliases = append(info.Aliases, e.Text)
	})

	// Images
	c.OnXML(`//*[@id="gallery"]/a`, func(e *colly.XMLElement) {
		if class := e.Attr("class"); class == "profile_img" || class == "profile_img_c" {
			return // ignore profile image here.
		}
		width := parser.ParseInt(e.Attr("data-width"))
		height := parser.ParseInt(e.Attr("data-height"))
		if width == 0 || height == 0 {
			return // width & height
		}
		info.Images = append(info.Images, e.Attr("href"))
	})

	// Images (profile)
	c.OnXML(`//img[@class="profile_img"]`, func(e *colly.XMLElement) {
		src := e.Attr("src")
		if strings.Trim(strings.TrimSpace(src), "#") == "" ||
			// ignore anonymous image:
			// https://xslist.org/assets/images/anonymous2.png
			strings.Contains(src, "anonymous") {
			return
		}
		info.Images = append(info.Images, e.Request.AbsoluteURL(src))
	})

	// Fields
	c.OnXML(`//*[@id="layout"]/div/p[1]`, func(e *colly.XMLElement) {
		for n := e.DOM.(*html.Node).FirstChild; n != nil; n = n.NextSibling {
			if n.Type != html.TextNode {
				continue
			}
			if ss := strings.SplitN(strings.TrimSpace(n.Data), ":", 2); len(ss) == 2 {
				if ss[1] = strings.TrimSpace(ss[1]); ss[1] == "" || ss[1] == "n/a" {
					continue
				}
				switch ss[0] {
				case "出生":
					info.Birthday = parser.ParseDate(ss[1])
				case "三围":
					info.Measurements = strings.ReplaceAll(ss[1], " ", "")
				case "罩杯":
					info.CupSize = strings.TrimSpace(strings.TrimSuffix(ss[1], "Cup"))
				case "出道日期":
					info.DebutDate = parseDebutDate(ss[1])
				case "血型":
					info.BloodType = ss[1]
				case "身高":
					info.Height = parser.ParseInt(strings.TrimRight(ss[1], "cm"))
				case "国籍":
					info.Nationality = ss[1]
				}
			}
		}
	})

	// Height
	c.OnXML(`//span[@itemprop="height"]`, func(e *colly.XMLElement) {
		info.Height = parser.ParseInt(strings.TrimRight(e.Text, "cm")) // ignore n/a
	})

	// Nationality
	c.OnXML(`//span[@itemprop="nationality"]`, func(e *colly.XMLElement) {
		info.Nationality = strings.ReplaceAll(e.Text, "n/a", "")
	})

	err = c.Visit(info.Homepage)
	return
}

func (xsl *XsList) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	c := xsl.ClonedCollector()

	c.OnXML(`//ul/li`, func(e *colly.XMLElement) {
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//h3/a`, "href"))
		id, _ := xsl.ParseActorIDFromURL(homepage)
		// Name
		actor := e.ChildAttr(`.//h3/a`, "title")
		if ss := strings.SplitN(actor, "-", 2); len(ss) == 2 {
			actor = strings.TrimSpace(ss[1])
		}
		// Images
		var images []string
		if img := e.ChildAttr(`.//div[1]/img`, "src"); img != "" {
			// NOTE: this might be an anonymous image link.
			// e.g.: https://xslist.org/assets/images/anonymous2.png
			images = []string{e.Request.AbsoluteURL(img)}
		}

		results = append(results, &model.ActorSearchResult{
			ID:       id,
			Name:     actor,
			Images:   images,
			Provider: xsl.Name(),
			Homepage: homepage,
		})
	})

	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func parseDebutDate(s string) dt.Date {
	if ss := regexp.MustCompile(`^([\s\d]+)年([\s\d]+)月$`).
		FindStringSubmatch(s); len(ss) == 3 {
		return dt.Date(time.Date(parser.ParseInt(ss[1]), time.Month(parser.ParseInt(ss[2])),
			2, 2, 2, 2, 2, time.UTC))
	}
	return parser.ParseDate(s)
}

func init() {
	provider.Register(Name, New)
}
