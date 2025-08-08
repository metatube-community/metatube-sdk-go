package modelmediaasia

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/collection/sets"
	"github.com/metatube-community/metatube-sdk-go/common/convertor"
	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

type ModelMediaAsiaActor struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func NewActorProvider() *ModelMediaAsiaActor {
	return &ModelMediaAsiaActor{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL}),
		Scraper: scraper.NewDefaultScraper(ActorProviderName, baseURL, Priority, language.Chinese),
	}
}

// ParseActorIDFromURL impls ActorProvider.ParseActorIDFromURL.
func (a *ModelMediaAsiaActor) ParseActorIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(homepage.Path), nil
}

// GetActorInfoByID impls ActorProvider.GetActorInfoByID.
func (a *ModelMediaAsiaActor) GetActorInfoByID(id string) (info *model.ActorInfo, err error) {
	info = &model.ActorInfo{
		ID:       id,
		Provider: a.Name(),
		Homepage: fmt.Sprintf(actorURL, id),
		Aliases:  []string{},
		Images:   []string{},
	}

	c := a.ClonedCollector()
	c.OnResponse(func(r *colly.Response) {
		resp := &actorInfoResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}

		// Name & Aliases
		info.Name = resp.Data.NameCn

		if resp.Data.Name != "" /* English name */ {
			info.Aliases = append(info.Aliases, resp.Data.Name)
		}

		// Images
		imageSet := sets.NewOrderedSet[string]()
		imageSet.Add(resp.Data.Avatar)
		for _, photo := range resp.Data.Photos {
			imageSet.Add(photo.Image)
		}
		info.Images = imageSet.AsSlice()

		// Birthday
		info.Birthday = parser.ParseDate(resp.Data.BirthDay)

		// Height
		if resp.Data.HeightCm > 0 {
			info.Height = resp.Data.HeightCm
		} else {
			info.Height = convertor.ConvertToCentimeters(
				resp.Data.HeightFt, resp.Data.HeightIn)
		}

		if size, cup, err := parser.ParseBustCupSize(resp.Data.MeasurementsChest); err == nil {
			if size != 0 &&
				resp.Data.MeasurementsWaist != 0 &&
				resp.Data.MeasurementsHips != 0 {
				info.Measurements = fmt.Sprintf("B:%d / W:%d / H:%d",
					size, resp.Data.MeasurementsWaist, resp.Data.MeasurementsHips)
			}
			info.CupSize = cup
		}
	})

	err = c.Visit(fmt.Sprintf(apiActorURL, id))
	return
}

// GetActorInfoByURL impls ActorProvider.GetActorInfoByURL.
func (a *ModelMediaAsiaActor) GetActorInfoByURL(rawURL string) (*model.ActorInfo, error) {
	id, err := a.ParseActorIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}

	return a.GetActorInfoByID(id)
}

// SearchActor impls ActorSearcher.SearchActor.
func (a *ModelMediaAsiaActor) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	c := a.ClonedCollector()

	c.OnResponse(func(r *colly.Response) {
		resp := &searchResponse{}
		if err = json.Unmarshal(r.Body, resp); err != nil {
			return
		}
		for _, actor := range resp.Data.Models {
			actorID := strconv.Itoa(actor.ID)
			res := &model.ActorSearchResult{
				ID:       actorID,
				Name:     actor.NameCn,
				Provider: a.Name(),
				Homepage: fmt.Sprintf(actorURL, actorID),
			}
			if actor.Avatar != "" {
				res.Images = append(res.Images, actor.Avatar)
			}
			if actor.Name != "" {
				res.Aliases = append(res.Aliases, actor.Name)
			}
			results = append(results, res)
		}
	})

	err = c.Visit(fmt.Sprintf(apiSearchURL, url.QueryEscape(keyword)))
	return
}
