package javdb

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider        = (*JavDB)(nil)
	_ provider.MovieSearcher        = (*JavDB)(nil)
	_ provider.ConfigSetter         = (*JavDB)(nil)
	_ provider.ProxySetter          = (*JavDB)(nil)
	_ provider.RequestTimeoutSetter = (*JavDB)(nil)
)

const (
	Name     = "JavDB"
	Priority = 1000 - 7

	baseURL          = "https://javdb.com/"
	defaultUserAgent = "curl/8.7.1"
)

var (
	videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	scorePattern   = regexp.MustCompile(`\d+(?:\.\d+)?`)
)

type JavDB struct {
	*scraper.Scraper

	cookie    string
	userAgent string
	client    *http.Client
}

func New() *JavDB {
	return &JavDB{
		Scraper: scraper.NewScraper(
			Name,
			baseURL,
			0,
			language.Chinese,
			scraper.WithAllowURLRevisit(),
			scraper.WithIgnoreRobotsTxt(),
			scraper.WithUserAgent(defaultUserAgent),
		),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (j *JavDB) SetRequestTimeout(timeout time.Duration) {
	j.Scraper.SetRequestTimeout(timeout)
	j.client.Timeout = timeout
}

func (j *JavDB) SetProxy(proxyURL string) error {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}
	if proxy.Scheme != "http" && proxy.Scheme != "https" {
		return fmt.Errorf("%s: unsupported proxy scheme %q", j.Name(), proxy.Scheme)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyURL(proxy)
	j.client.Transport = transport
	return j.Scraper.SetProxy(proxyURL)
}

func (j *JavDB) SetConfig(config provider.Config) error {
	j.cookie = ""
	j.userAgent = defaultUserAgent
	j.SetPriority(0)

	if !config.Has("cookie") {
		return nil
	}

	cookie, err := config.GetString("cookie")
	if err != nil {
		return err
	}
	if cookie = strings.TrimSpace(cookie); cookie == "" {
		return nil
	}

	j.cookie = cookie
	if config.Has("user_agent") {
		userAgent, err := config.GetString("user_agent")
		if err != nil {
			return err
		}
		if userAgent = strings.TrimSpace(userAgent); userAgent != "" {
			j.userAgent = userAgent
		}
	}
	j.SetPriority(Priority)
	return nil
}

func (j *JavDB) NormalizeMovieID(id string) string {
	id = strings.TrimSpace(id)
	if !videoIDPattern.MatchString(id) {
		return ""
	}
	return id
}

func (j *JavDB) ParseMovieIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(u.Hostname(), j.URL().Hostname()) {
		return "", provider.ErrInvalidURL
	}
	id := j.NormalizeMovieID(path.Base(u.Path))
	if id == "" || !strings.HasPrefix(u.Path, "/v/") {
		return "", provider.ErrInvalidID
	}
	return id, nil
}

func (j *JavDB) GetMovieInfoByID(id string) (*model.MovieInfo, error) {
	if id = j.NormalizeMovieID(id); id == "" {
		return nil, provider.ErrInvalidID
	}
	return j.GetMovieInfoByURL(j.absoluteURL("/v/" + id))
}

func (j *JavDB) GetMovieInfoByURL(rawURL string) (*model.MovieInfo, error) {
	id, err := j.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}

	body, err := j.get(rawURL)
	if err != nil {
		return nil, err
	}
	return j.parseMovieInfo(body, id, rawURL)
}

func (j *JavDB) NormalizeMovieKeyword(keyword string) string {
	return strings.ToUpper(number.Trim(keyword))
}

func (j *JavDB) SearchMovie(keyword string) ([]*model.MovieSearchResult, error) {
	if keyword = j.NormalizeMovieKeyword(keyword); keyword == "" {
		return nil, provider.ErrInvalidKeyword
	}

	body, err := j.get(j.absoluteURL("/search?q=" + url.QueryEscape(keyword) + "&f=all&locale=zh"))
	if err != nil {
		return nil, err
	}
	return j.parseSearchResults(body)
}

func (j *JavDB) get(rawURL string) ([]byte, error) {
	if j.cookie == "" {
		return nil, fmt.Errorf("%s: cookie is not configured", j.Name())
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", j.cookie)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("User-Agent", j.userAgent)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", j.Name(), err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response: %w", j.Name(), err)
	}
	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, provider.ErrInfoNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("%s: authentication rejected (HTTP %d)", j.Name(), resp.StatusCode)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s: unexpected HTTP status %d", j.Name(), resp.StatusCode)
	}
	if isChallenge(body) {
		return nil, fmt.Errorf("%s: anti-bot challenge response", j.Name())
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("%s: empty response", j.Name())
	}
	return body, nil
}

func (j *JavDB) parseSearchResults(body []byte) ([]*model.MovieSearchResult, error) {
	if isChallenge(body) {
		return nil, fmt.Errorf("%s: anti-bot challenge response", j.Name())
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%s: parse search response: %w", j.Name(), err)
	}
	items := doc.Find("div.item")
	if items.Length() == 0 {
		if doc.Find(".empty-message").Length() > 0 {
			return nil, provider.ErrInfoNotFound
		}
		return nil, fmt.Errorf("%s: unrecognized search response", j.Name())
	}

	results := make([]*model.MovieSearchResult, 0, items.Length())
	items.Each(func(_ int, item *goquery.Selection) {
		href, ok := item.Find(".box").Attr("href")
		if !ok {
			return
		}
		id := j.movieIDFromPath(href)
		if id == "" {
			return
		}
		titleSelection := item.Find(".video-title")
		number := strings.ToUpper(strings.TrimSpace(titleSelection.Find("strong").Text()))
		title := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(titleSelection.Text()), number))
		if number == "" || title == "" {
			return
		}
		cover, _ := item.Find(".cover img").Attr("src")
		result := &model.MovieSearchResult{
			ID:       id,
			Number:   number,
			Title:    title,
			Provider: j.Name(),
			Homepage: j.absoluteURL(href),
			ThumbURL: j.absoluteURL(cover),
			CoverURL: j.absoluteURL(cover),
			Score:    parseScore(item.Find(".score").Text()),
		}
		if date := strings.TrimSpace(item.Find(".meta").Text()); date != "" {
			result.ReleaseDate = parser.ParseDate(date)
		}
		results = append(results, result)
	})
	if len(results) == 0 {
		return nil, fmt.Errorf("%s: search page has no valid results", j.Name())
	}
	return results, nil
}

func (j *JavDB) parseMovieInfo(body []byte, id, homepage string) (*model.MovieInfo, error) {
	if isChallenge(body) {
		return nil, fmt.Errorf("%s: anti-bot challenge response", j.Name())
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%s: parse detail response: %w", j.Name(), err)
	}

	info := &model.MovieInfo{
		ID:            id,
		Provider:      j.Name(),
		Homepage:      homepage,
		Title:         strings.TrimSpace(doc.Find(".current-title").First().Text()),
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}
	cover, _ := doc.Find(".video-cover").First().Attr("src")
	info.CoverURL = j.absoluteURL(cover)
	info.ThumbURL = info.CoverURL

	doc.Find(".movie-panel-info .panel-block").Each(func(_ int, block *goquery.Selection) {
		label := strings.Trim(strings.TrimSpace(block.Find("strong").First().Text()), ":：")
		value := block.Find(".value").First()
		switch label {
		case "番號", "番号":
			if number, ok := value.Find("[data-clipboard-text]").Attr("data-clipboard-text"); ok {
				info.Number = strings.ToUpper(strings.TrimSpace(number))
			} else {
				info.Number = strings.ToUpper(strings.TrimSpace(value.Text()))
			}
		case "日期":
			info.ReleaseDate = parser.ParseDate(value.Text())
		case "時長", "时长":
			info.Runtime = parser.ParseRuntime(value.Text())
		case "評分", "评分":
			info.Score = parseScore(value.Text())
		case "導演", "导演":
			info.Director = strings.TrimSpace(value.Find("a").First().Text())
		case "片商":
			info.Maker = strings.TrimSpace(value.Find("a").First().Text())
		case "系列":
			info.Series = strings.TrimSpace(value.Find("a").First().Text())
		case "類別", "类别":
			value.Find("a").Each(func(_ int, genre *goquery.Selection) {
				if name := strings.TrimSpace(genre.Text()); name != "" {
					info.Genres = append(info.Genres, name)
				}
			})
		case "演員", "演员":
			value.Find("a").Each(func(_ int, actor *goquery.Selection) {
				if name := strings.TrimSpace(actor.Text()); name != "" {
					info.Actors = append(info.Actors, name)
				}
			})
		}
	})
	doc.Find(".preview-images .tile-item").Each(func(_ int, image *goquery.Selection) {
		if href, ok := image.Attr("href"); ok {
			info.PreviewImages = append(info.PreviewImages, j.absoluteURL(href))
		}
	})
	if info.Title == "" || info.Number == "" {
		return nil, fmt.Errorf("%s: unrecognized detail response", j.Name())
	}
	return info, nil
}

func (j *JavDB) movieIDFromPath(href string) string {
	u, err := url.Parse(href)
	if err != nil || !strings.HasPrefix(u.Path, "/v/") {
		return ""
	}
	return j.NormalizeMovieID(path.Base(u.Path))
}

func (j *JavDB) absoluteURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return j.URL().ResolveReference(u).String()
}

func parseScore(raw string) float64 {
	match := scorePattern.FindString(strings.TrimSpace(raw))
	score, _ := strconv.ParseFloat(match, 64)
	return score
}

func isChallenge(body []byte) bool {
	body = bytes.ToLower(body)
	return bytes.Contains(body, []byte("cf-chl")) ||
		bytes.Contains(body, []byte("challenge-platform")) ||
		bytes.Contains(body, []byte("just a moment"))
}

func init() {
	provider.Register(Name, New)
}
