package theporndb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.ActorProvider = (*ThePornDBActor)(nil)
	_ provider.ActorSearcher = (*ThePornDBActor)(nil)
)

const (
	ActorProviderName = "ThePornDBActor"
	actorBaseURL      = "https://theporndb.net/performers/"
	actorPageURL      = "https://theporndb.net/performers/%s"
	apiGetActorURL    = "https://api.theporndb.net/performers/%s"
	apiSearchActorURL = "https://api.theporndb.net/performers?q=%s"
)

type ThePornDBActor struct {
	*scraper.Scraper

	accessToken string
}

func NewThePornDBActor() *ThePornDBActor {
	return &ThePornDBActor{
		Scraper:     scraper.NewDefaultScraper(ActorProviderName, actorBaseURL, Priority, language.English),
		accessToken: "",
	}
}

func (s *ThePornDBActor) SetConfig(config map[string]string) error {
	if accessToken, ok := config["ACCESS_TOKEN"]; ok {
		s.accessToken = accessToken
	}
	return nil
}

// ParseActorIDFromURL impls ActorProvider.ParseActorIDFromURL.
func (s *ThePornDBActor) ParseActorIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

// GetActorInfoByID impls ActorProvider.GetActorInfoByID.
func (s *ThePornDBActor) GetActorInfoByID(id string) (info *model.ActorInfo, err error) {
	if s.accessToken == "" {
		return nil, nil
	}

	info = &model.ActorInfo{
		Provider: s.Name(),
		Aliases:  []string{},
		Images:   []string{},
	}

	c := s.ClonedCollector()
	c.OnResponse(func(r *colly.Response) {
		resp := &getActorResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}

		info.ID = resp.Data.Slug
		info.Homepage = fmt.Sprintf(actorPageURL, info.ID)
		info.Name = resp.Data.Name
		info.Summary = resp.Data.Bio
		info.Nationality = resp.Data.Extras.Nationality
		info.Aliases = resp.Data.Aliases

		if birthday, err := resp.Data.BirthdayDate(); err == nil {
			info.Birthday = birthday
		}

		if resp.Data.Image != "" {
			info.Images = append(info.Images, resp.Data.Image)
		}
		if resp.Data.Thumbnail != "" {
			info.Images = append(info.Images, resp.Data.Thumbnail)
		}
		if resp.Data.Face != "" {
			info.Images = append(info.Images, resp.Data.Face)
		}
		for _, poster := range resp.Data.Posters {
			info.Images = append(info.Images, poster.URL)
		}

		info.Height = resp.Data.HeightInCM()

		b, cup, err := parseChestSize(resp.Data.Extras.Cupsize)
		if err == nil {
			info.CupSize = cup

			w, _ := strconv.Atoi(resp.Data.Extras.Waist)
			h, _ := strconv.Atoi(resp.Data.Extras.Hips)

			if b != 0 && w != 0 && h != 0 {
				info.Measurements = fmt.Sprintf("B:%d / W:%d / H:%d", b, w, h)
			}
		}
	})

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	err = c.Request(http.MethodGet, fmt.Sprintf(apiGetActorURL, id), nil, nil, headers)
	return
}

// GetActorInfoByURL impls ActorProvider.GetActorInfoByURL.
func (s *ThePornDBActor) GetActorInfoByURL(rawURL string) (*model.ActorInfo, error) {
	id, err := s.ParseActorIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}

	return s.GetActorInfoByID(id)
}

// SearchActor impls ActorSearcher.SearchActor.
func (s *ThePornDBActor) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	if s.accessToken == "" {
		return nil, nil
	}

	c := s.ClonedCollector()

	results = make([]*model.ActorSearchResult, 0)

	c.OnResponse(func(r *colly.Response) {
		resp := &searchActorResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}
		for _, actor := range resp.Data {
			res := &model.ActorSearchResult{
				ID:       actor.Slug,
				Name:     actor.Name,
				Provider: s.Name(),
				Homepage: fmt.Sprintf(actorPageURL, actor.Slug),
			}
			res.Aliases = actor.Aliases

			if actor.Image != "" {
				res.Images = append(res.Images, actor.Image)
			}
			if actor.Thumbnail != "" {
				res.Images = append(res.Images, actor.Thumbnail)
			}
			if actor.Face != "" {
				res.Images = append(res.Images, actor.Face)
			}
			for _, poster := range actor.Posters {
				res.Images = append(res.Images, poster.URL)
			}
			results = append(results, res)
		}
	})

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", s.accessToken))
	err = c.Request(http.MethodGet, fmt.Sprintf(apiSearchActorURL, url.QueryEscape(keyword)), nil, nil, headers)
	return
}

var chestSizeRE = regexp.MustCompile(`^(\d+)([A-Z])$`)

func parseChestSize(s string) (int, string, error) {
	match := chestSizeRE.FindStringSubmatch(s)

	if len(match) != 3 {
		return 0, "", fmt.Errorf("invalid format: %s", s)
	}

	numericPart := match[1]
	unitPart := match[2]

	value, err := strconv.Atoi(numericPart)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse numeric part '%s': %w", numericPart, err)
	}

	return value, unitPart, nil
}
