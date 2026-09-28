package fantia

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

const (
	PostName    = "FantiaPost"
	postBaseURL = baseURL + "posts/"
	postAPIURL  = baseURL + "api/v1/posts/%s"
)

var postNumericID = regexp.MustCompile(`^[1-9]\d*$`)

var _ provider.MovieProvider = (*Post)(nil)
var _ provider.ConfigSetter = (*Post)(nil)

type Post struct {
	*scraper.Scraper
	sessionID string
}

func NewPost() *Post {
	return &Post{
		Scraper: scraper.NewDefaultScraper(PostName, postBaseURL, Priority, language.Japanese,
			scraper.WithCookies(baseURL, []*http.Cookie{{Name: "age_check", Value: "1"}})),
		sessionID: strings.TrimSpace(os.Getenv("FANTIA_SESSION_ID")),
	}
}

func (p *Post) SetConfig(c provider.Config) error {
	if !c.Has("session_id") {
		return nil
	}
	v, err := c.GetString("session_id")
	if err != nil {
		return err
	}
	p.sessionID = strings.TrimSpace(v)
	return nil
}

func (p *Post) NormalizeMovieID(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	id = strings.NewReplacer("-", "", "_", "", " ", "").Replace(id)
	id = strings.TrimPrefix(id, "FANTIAPOSTS")
	id = strings.TrimPrefix(id, "FANTIAPOST")
	if !postNumericID.MatchString(id) {
		return ""
	}
	return id
}

func (p *Post) ParseMovieIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, "fantia.jp") {
		return "", provider.ErrInvalidURL
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "posts" {
		return "", provider.ErrInvalidURL
	}
	id := p.NormalizeMovieID(parts[1])
	if id == "" {
		return "", provider.ErrInvalidID
	}
	return id, nil
}

func (p *Post) GetMovieInfoByID(id string) (*model.MovieInfo, error) {
	id = p.NormalizeMovieID(id)
	if id == "" {
		return nil, provider.ErrInvalidID
	}
	return p.getPostInfo(id)
}

func (p *Post) GetMovieInfoByURL(rawURL string) (*model.MovieInfo, error) {
	id, err := p.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}
	return p.getPostInfo(id)
}

func (p *Post) getPostInfo(id string) (*model.MovieInfo, error) {
	c := p.ClonedCollector()
	if p.sessionID != "" {
		if err := c.SetCookies(baseURL, []*http.Cookie{{
			Name: "_session_id", Value: p.sessionID, Path: "/", Secure: true, HttpOnly: true,
		}}); err != nil {
			return nil, fmt.Errorf("set Fantia session cookie: %w", err)
		}
	}

	var csrfToken string
	c.OnHTML(`meta[name="csrf-token"]`, func(e *colly.HTMLElement) {
		csrfToken = strings.TrimSpace(e.Attr("content"))
	})
	if err := c.Visit(baseURL); err != nil {
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
	headers.Set("Accept", "application/json")
	headers.Set("X-CSRF-Token", csrfToken)
	headers.Set("X-Requested-With", "XMLHttpRequest")
	if err := c.Request(http.MethodGet, fmt.Sprintf(postAPIURL, id), nil, nil, headers); err != nil {
		return nil, err
	}
	if parseErr != nil {
		return nil, parseErr
	}
	if strconv.FormatInt(response.Post.ID, 10) != id {
		return nil, provider.ErrInfoNotFound
	}
	info := p.postMovieInfo(id, &response.Post)
	if !info.IsValid() {
		return nil, provider.ErrIncompleteMetadata
	}
	return info, nil
}

func (p *Post) postMovieInfo(id string, post *postData) *model.MovieInfo {
	info := &model.MovieInfo{
		ID:            id,
		Number:        "FANTIA-POST-" + id,
		Title:         strings.TrimSpace(post.Title),
		Summary:       strings.TrimSpace(post.Comment),
		Provider:      p.Name(),
		Homepage:      postBaseURL + id,
		Maker:         strings.TrimSpace(post.Fanclub.Name),
		Label:         strings.TrimSpace(post.Rating),
		ReleaseDate:   parser.ParseDate(post.PostedAt),
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}
	if creator := strings.TrimSpace(post.Fanclub.User.Name); creator != "" {
		info.Actors = append(info.Actors, creator)
	}
	for _, tag := range post.Tags {
		info.Genres = appendUnique(info.Genres, strings.TrimSpace(tag.Name))
	}

	cover := firstURL(post.Thumb.Original, post.Thumb.Main,
		post.Fanclub.Cover.Original, post.Fanclub.Cover.Main,
		post.Fanclub.Cover.OGP, post.Fanclub.Icon.Original)
	var blogTexts []string
	for _, content := range post.PostContents {
		if content.VisibleStatus != "" && content.VisibleStatus != "visible" {
			continue
		}
		for _, photo := range content.Photos {
			info.PreviewImages = appendUnique(info.PreviewImages,
				firstURL(photo.URL.Original, photo.URL.Main, photo.URL.Medium))
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
			body, images := parseBlog(content.Comment)
			if body != "" {
				blogTexts = append(blogTexts, body)
			}
			for _, image := range images {
				info.PreviewImages = appendUnique(info.PreviewImages, absoluteFantiaURL(image))
			}
		}
	}
	if body := strings.Join(blogTexts, "\n\n"); body != "" {
		switch {
		case info.Summary == "":
			info.Summary = body
		case strings.Contains(body, info.Summary):
			info.Summary = body
		case !strings.Contains(info.Summary, body):
			info.Summary += "\n\n" + body
		}
	}
	if cover == "" && len(info.PreviewImages) > 0 {
		cover = info.PreviewImages[0]
	}
	info.ThumbURL = cover
	info.CoverURL = cover
	return info
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

type fantiaImage struct {
	Thumb    string `json:"thumb"`
	Medium   string `json:"medium"`
	Main     string `json:"main"`
	Original string `json:"original"`
	OGP      string `json:"ogp"`
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
		if json.Unmarshal(op.Insert, &image) == nil {
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
	base, _ := url.Parse(baseURL)
	reference, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(reference)
	if resolved.Scheme != "https" && resolved.Scheme != "http" {
		return ""
	}
	return resolved.String()
}

func firstURL(values ...string) string {
	for _, value := range values {
		if resolved := absoluteFantiaURL(value); resolved != "" {
			return resolved
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
