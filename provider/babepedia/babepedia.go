package babepedia

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"
	"gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

const (
	Name         = "Babepedia"
	actorBaseURL = "https://www.babepedia.com"
	actorPageURL = "https://www.babepedia.com/babe/%s"
	priority     = 1000
)

var (
	_ provider.ActorProvider = (*BabepediaActor)(nil)
	_ provider.ActorSearcher = (*BabepediaActor)(nil)
)

type BabepediaActor struct {
	*scraper.Scraper
}

func New() *BabepediaActor {
	return &BabepediaActor{
		Scraper: scraper.NewDefaultScraper(
			Name,
			actorBaseURL,
			priority, // Priority
			language.English,
			scraper.WithHeaders(map[string]string{
				"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
				"Accept-Language":           "en-US,en;q=0.5",
				"Accept-Encoding":           "gzip, deflate",
				"Connection":                "keep-alive",
				"Upgrade-Insecure-Requests": "1",
			}),
		),
	}
}

func (b *BabepediaActor) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	encodedName := strings.ReplaceAll(keyword, " ", "_")

	pageURL := fmt.Sprintf(actorPageURL, encodedName)

	// Create a new collector
	c := b.ClonedCollector()

	results = make([]*model.ActorSearchResult, 0)

	c.OnHTML(".results", func(e *colly.HTMLElement) {
		e.ForEach(".thumbshot", func(_ int, el *colly.HTMLElement) {
			u := e.ChildAttr("a", "href")
			if u != "" {
				absoluteURL := e.Request.AbsoluteURL(u)
				d := c.Clone()
				result := &model.ActorSearchResult{
					Provider: Name,
				}
				d.OnHTML("#profile-info", func(e *colly.HTMLElement) {
					result.ID, _ = b.ParseActorIDFromURL(absoluteURL)
					// Extract actor name from h1#babename
					result.Name = e.ChildText("#babename")

					// Set homepage to current page URL
					result.Homepage = e.Request.URL.String()

					// Extract aliases from #aliasinfo .aliasbox elements
					e.ForEach("#aliasinfo .aliasbox", func(_ int, el *colly.HTMLElement) {
						aliasName := el.ChildText(".aliasname")
						if aliasName != "" && aliasName != result.Name {
							result.Aliases = append(result.Aliases, aliasName)
						}
					})

					if result.Name != "" {
						results = append(results, result)
					}
				})

				d.Visit(absoluteURL)
			}

		})
	})

	c.Visit(pageURL)

	return
}

func (b *BabepediaActor) ParseActorIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

func (b *BabepediaActor) GetActorInfoByID(id string) (info *model.ActorInfo, err error) {
	pageURL := fmt.Sprintf(actorPageURL, id)

	// Create a new collector
	c := b.ClonedCollector()
	info = &model.ActorInfo{
		Provider: Name,
		ID:       id,
	}

	c.OnHTML("#profile-info", func(e *colly.HTMLElement) {
		info.Homepage = e.Request.URL.String()
		name := e.ChildText("#babename")
		if name != "" {
			info.Name = name
		}
	})

	c.OnHTML("#biotext", func(e *colly.HTMLElement) {
		info.Summary = e.Text
	})

	// Extract personal info including Birthplace, Born, and Nationality
	c.OnHTML(".info-grid", func(e *colly.HTMLElement) {
		e.ForEach(".info-item", func(_ int, el *colly.HTMLElement) {
			label := el.ChildText(".label")
			value := el.ChildText(".value")

			// Clean up the value by removing extra whitespace
			value = strings.TrimSpace(value)

			switch label {
			case "Birthplace:":
				info.Hobby = value // Using Hobby field to store Birthplace
			case "Born:":
				// Try to parse the birth date
				// The format is like "Tuesday 19th of December 2000"
				// We'll extract just the date part and try to parse it
				re := regexp.MustCompile(`(\d{1,2})(st|nd|rd|th) of (\w+) (\d{4})`)
				matches := re.FindStringSubmatch(value)
				if len(matches) == 5 {
					dateStr := fmt.Sprintf("%s %s %s", matches[1], matches[3], matches[4])
					if t, err := time.Parse("2 January 2006", dateStr); err == nil {
						info.Birthday = datatypes.Date(t)
					}
				}
			case "Nationality:":
				// Clean up nationality value (just remove parentheses)
				re := regexp.MustCompile(`[()]`)
				info.Nationality = strings.TrimSpace(re.ReplaceAllString(value, ""))
			}
		})
	})

	c.OnHTML(".thumbnail", func(e *colly.HTMLElement) {
		u := e.ChildAttr("a", "href")
		if u != "" {
			absoluteURL := e.Request.AbsoluteURL(u)
			info.Images = append(info.Images, absoluteURL)
		}
	})

	c.Visit(pageURL)
	return
}

func (b *BabepediaActor) GetActorInfoByURL(rawURL string) (*model.ActorInfo, error) {
	id, err := b.ParseActorIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}
	return b.GetActorInfoByID(id)
}

func init() {
	provider.Register(Name, New)
}
