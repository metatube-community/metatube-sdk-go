package fantia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*Fantia)(nil)
	_ provider.ConfigSetter  = (*Fantia)(nil)
)

const (
	PostProviderName    = "FantiaPost"
	ProductProviderName = "FantiaProduct"
	Priority            = 1000

	rootURL    = "https://fantia.jp"
	postAPIURL = rootURL + "/api/v1/posts/%s"
)

var numericID = regexp.MustCompile(`^[1-9]\d*$`)

type Fantia struct {
	*scraper.Scraper

	kind      string
	pageURL   string
	sessionID string
}

func NewPost() *Fantia {
	return newFantia(PostProviderName, "post")
}

func NewProduct() *Fantia {
	return newFantia(ProductProviderName, "product")
}

func newFantia(name, kind string) *Fantia {
	baseURL := fmt.Sprintf("%s/%ss/", rootURL, kind)
	return &Fantia{
		Scraper:   scraper.NewDefaultScraper(name, baseURL, Priority, language.Japanese),
		kind:      kind,
		pageURL:   baseURL + "%s",
		sessionID: strings.TrimSpace(os.Getenv("FANTIA_SESSION_ID")),
	}
}

func (f *Fantia) SetConfig(config provider.Config) error {
	if !config.Has("session_id") {
		return nil
	}
	sessionID, err := config.GetString("session_id")
	if err != nil {
		return err
	}
	f.sessionID = strings.TrimSpace(sessionID)
	return nil
}

func (f *Fantia) NormalizeMovieID(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	id = strings.NewReplacer("-", "", "_", "", " ", "").Replace(id)
	prefix := "FANTIA" + strings.ToUpper(f.kind)
	id = strings.TrimPrefix(id, prefix+"S")
	id = strings.TrimPrefix(id, prefix)
	if !numericID.MatchString(id) {
		return ""
	}
	return id
}

func (f *Fantia) ParseMovieIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Hostname() != "fantia.jp" && u.Hostname() != "www.fantia.jp") {
		return "", provider.ErrInvalidURL
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != f.kind+"s" {
		return "", provider.ErrInvalidURL
	}
	id := f.NormalizeMovieID(parts[1])
	if id == "" {
		return "", provider.ErrInvalidID
	}
	return id, nil
}

func (f *Fantia) GetMovieInfoByID(id string) (*model.MovieInfo, error) {
	id = f.NormalizeMovieID(id)
	if id == "" {
		return nil, provider.ErrInvalidID
	}
	if f.kind == "post" {
		return f.getPostInfo(id)
	}
	return f.getProductInfo(id)
}

func (f *Fantia) GetMovieInfoByURL(rawURL string) (*model.MovieInfo, error) {
	id, err := f.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}
	return f.GetMovieInfoByID(id)
}

func (f *Fantia) collector() (*colly.Collector, error) {
	c := f.ClonedCollector()
	if f.sessionID == "" {
		return c, nil
	}
	err := c.SetCookies(rootURL, []*http.Cookie{{
		Name:     "_session_id",
		Value:    f.sessionID,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}})
	return c, err
}

func (f *Fantia) newMovieInfo(id string) *model.MovieInfo {
	return &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("FANTIA-%s-%s", strings.ToUpper(f.kind), id),
		Provider:      f.Name(),
		Homepage:      fmt.Sprintf(f.pageURL, id),
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}
}

func (f *Fantia) getPostInfo(id string) (*model.MovieInfo, error) {
	c, err := f.collector()
	if err != nil {
		return nil, err
	}

	csrfToken := ""
	c.OnHTML(`meta[name="csrf-token"]`, func(e *colly.HTMLElement) {
		csrfToken = strings.TrimSpace(e.Attr("content"))
	})
	if err := c.Visit(rootURL + "/"); err != nil {
		return nil, err
	}
	if csrfToken == "" {
		return nil, provider.ErrInfoNotFound
	}

	var response postResponse
	var parseErr error
	c.OnResponse(func(r *colly.Response) {
		parseErr = json.Unmarshal(r.Body, &response)
	})
	headers := http.Header{}
	headers.Set("Accept", "application/json, text/plain, */*")
	headers.Set("X-CSRF-Token", csrfToken)
	headers.Set("X-Requested-With", "XMLHttpRequest")
	if err := c.Request(http.MethodGet, fmt.Sprintf(postAPIURL, id), nil, nil, headers); err != nil {
		return nil, err
	}
	if parseErr != nil {
		return nil, parseErr
	}
	if response.Post.ID == 0 {
		return nil, provider.ErrInfoNotFound
	}
	return f.postMovieInfo(id, &response.Post), nil
}

func (f *Fantia) postMovieInfo(id string, post *postData) *model.MovieInfo {
	info := f.newMovieInfo(id)
	info.Title = strings.TrimSpace(post.Title)
	info.Summary = strings.TrimSpace(post.Comment)
	info.Label = strings.TrimSpace(post.Rating)
	info.Maker = strings.TrimSpace(post.Fanclub.Name)
	info.ReleaseDate = parser.ParseDate(post.PostedAt)
	if creator := strings.TrimSpace(post.Fanclub.User.Name); creator != "" {
		info.Actors = append(info.Actors, creator)
	}
	for _, tag := range post.Tags {
		if name := strings.TrimSpace(tag.Name); name != "" {
			info.Genres = appendUnique(info.Genres, name)
		}
	}

	cover := firstURL(
		post.Thumb.Original,
		post.Fanclub.Cover.Original,
		post.Fanclub.Cover.Main,
		post.Fanclub.Cover.OGP,
		post.Fanclub.Icon.Original,
		post.Fanclub.Icon.Main,
	)
	for _, content := range post.PostContents {
		if content.VisibleStatus != "visible" {
			continue
		}
		for _, photo := range content.Photos {
			info.PreviewImages = appendUnique(info.PreviewImages, absoluteFantiaURL(photo.URL.Original))
		}
		if downloadURL := absoluteFantiaURL(content.DownloadURI); downloadURL != "" {
			switch strings.ToLower(path.Ext(urlPath(downloadURL))) {
			case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif":
				info.PreviewImages = appendUnique(info.PreviewImages, downloadURL)
			case ".mp4", ".m4v", ".mov", ".webm":
				if info.PreviewVideoURL == "" {
					info.PreviewVideoURL = downloadURL
				}
			}
		}
		if content.Category == "blog" {
			text, images := parseBlog(content.Comment)
			if info.Summary == "" {
				info.Summary = text
			}
			for _, image := range images {
				info.PreviewImages = appendUnique(info.PreviewImages, absoluteFantiaURL(image))
			}
		}
	}
	if cover == "" && len(info.PreviewImages) > 0 {
		cover = info.PreviewImages[0]
	}
	info.ThumbURL = cover
	info.CoverURL = cover
	return info
}

func (f *Fantia) getProductInfo(id string) (*model.MovieInfo, error) {
	info := f.newMovieInfo(id)
	c, err := f.collector()
	if err != nil {
		return nil, err
	}

	c.OnHTML(`meta[property="og:title"]`, func(e *colly.HTMLElement) {
		info.Title = strings.TrimSpace(e.Attr("content"))
	})
	c.OnHTML(`meta[property="og:description"]`, func(e *colly.HTMLElement) {
		info.Summary = strings.TrimSpace(e.Attr("content"))
	})
	c.OnHTML(`meta[property="og:image"]`, func(e *colly.HTMLElement) {
		image := absoluteFantiaURL(e.Attr("content"))
		info.ThumbURL = image
		info.CoverURL = image
	})
	c.OnHTML(`script[type="application/ld+json"]`, func(e *colly.HTMLElement) {
		for _, product := range decodeProducts(e.Text) {
			if product.Type != "Product" {
				continue
			}
			if product.Name != "" {
				info.Title = strings.TrimSpace(product.Name)
			}
			if product.Description != "" {
				info.Summary = strings.TrimSpace(product.Description)
			}
			if product.Brand.Name != "" {
				info.Maker = strings.TrimSpace(product.Brand.Name)
			}
			if product.Category != "" {
				info.Genres = appendUnique(info.Genres, strings.TrimSpace(product.Category))
			}
			for i, rawImage := range product.Image {
				image := absoluteFantiaURL(rawImage)
				if image == "" {
					continue
				}
				if i == 0 {
					info.ThumbURL = image
					info.CoverURL = image
				} else {
					info.PreviewImages = appendUnique(info.PreviewImages, image)
				}
			}
		}
	})
	c.OnHTML(`h1.product-title`, func(e *colly.HTMLElement) {
		if title := strings.TrimSpace(e.Text); title != "" {
			info.Title = title
		}
	})
	c.OnHTML(`.fanclub-show-header h1.fanclub-name a`, func(e *colly.HTMLElement) {
		if info.Maker == "" {
			info.Maker = strings.TrimSpace(e.Text)
		}
	})

	if err := c.Visit(info.Homepage + "?locale=jp"); err != nil {
		return nil, err
	}
	if info.Title == "" {
		return nil, provider.ErrInfoNotFound
	}
	return info, nil
}

type postResponse struct {
	Post postData `json:"post"`
}

type postData struct {
	ID       int64       `json:"id"`
	Title    string      `json:"title"`
	Comment  string      `json:"comment"`
	Rating   string      `json:"rating"`
	PostedAt string      `json:"posted_at"`
	Thumb    fantiaImage `json:"thumb"`
	Fanclub  struct {
		Name  string      `json:"name"`
		Cover fantiaImage `json:"cover"`
		Icon  fantiaImage `json:"icon"`
		User  struct {
			Name string `json:"name"`
		} `json:"user"`
	} `json:"fanclub"`
	Tags         []postTag     `json:"tags"`
	PostContents []postContent `json:"post_contents"`
}

type postTag struct {
	Name string `json:"name"`
}

type postContent struct {
	Category      string      `json:"category"`
	Comment       string      `json:"comment"`
	DownloadURI   string      `json:"download_uri"`
	VisibleStatus string      `json:"visible_status"`
	Photos        []postPhoto `json:"post_content_photos"`
}

type postPhoto struct {
	URL fantiaImage `json:"url"`
}

type schemaProduct struct {
	Type        string     `json:"@type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Image       stringList `json:"image"`
	Brand       struct {
		Name string `json:"name"`
	} `json:"brand"`
}

type fantiaImage struct {
	Thumb    string `json:"thumb"`
	Medium   string `json:"medium"`
	Main     string `json:"main"`
	Original string `json:"original"`
	OGP      string `json:"ogp"`
}

type stringList []string

func (s *stringList) UnmarshalJSON(data []byte) error {
	var values []string
	if err := json.Unmarshal(data, &values); err == nil {
		*s = values
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = []string{value}
	return nil
}

func decodeProducts(raw string) []schemaProduct {
	data := bytes.TrimSpace([]byte(raw))
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		var products []schemaProduct
		_ = json.Unmarshal(data, &products)
		return products
	}
	var product schemaProduct
	if json.Unmarshal(data, &product) != nil {
		return nil
	}
	return []schemaProduct{product}
}

func parseBlog(raw string) (string, []string) {
	var delta struct {
		Ops []struct {
			Insert json.RawMessage `json:"insert"`
		} `json:"ops"`
	}
	if json.Unmarshal([]byte(raw), &delta) != nil {
		return "", nil
	}

	var text strings.Builder
	var images []string
	for _, op := range delta.Ops {
		var value string
		if json.Unmarshal(op.Insert, &value) == nil {
			text.WriteString(value)
			continue
		}
		var image struct {
			FantiaImage struct {
				OriginalURL string `json:"original_url"`
			} `json:"fantiaImage"`
		}
		if json.Unmarshal(op.Insert, &image) == nil && image.FantiaImage.OriginalURL != "" {
			images = appendUnique(images, image.FantiaImage.OriginalURL)
		}
	}
	return strings.TrimSpace(text.String()), images
}

func absoluteFantiaURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	base, _ := url.Parse(rootURL)
	reference, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return base.ResolveReference(reference).String()
}

func firstURL(values ...string) string {
	for _, value := range values {
		if value = absoluteFantiaURL(value); value != "" {
			return value
		}
	}
	return ""
}

func urlPath(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Path
}

func appendUnique(values []string, value string) []string {
	if value == "" {
		return values
	}
	for _, current := range values {
		if current == value {
			return values
		}
	}
	return append(values, value)
}

func init() {
	provider.Register(PostProviderName, NewPost)
	provider.Register(ProductProviderName, NewProduct)
}
