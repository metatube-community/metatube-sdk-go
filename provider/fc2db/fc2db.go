package fc2db

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fc2/fc2util"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*FC2DB)(nil)
	_ provider.MovieSearcher = (*FC2DB)(nil)
)

const (
	Name     = "FC2DB"
	Priority = 1000 - 3
)

const (
	baseURL   = "https://fc2db.net/"
	movieURL  = "https://fc2db.net/work/%s/"
	searchURL = "https://fc2db.net/?s=%s"
	// FC2 官方接口：{"path":"...mp4","poster_image_path":"...横版剧照","code":200}
	sampleAPIURL = "https://adult.contents.fc2.com/api/v2/videos/%s/sample"
)

// h1 on work page looks like: [FC2-PPV-4974437] Title...
var titlePattern = regexp.MustCompile(`^\[\s*FC2[-_\s]?PPV[-_\s]?(\d+)\s*\]\s*(.*)$`)

type FC2DB struct {
	*scraper.Scraper
}

func New() *FC2DB {
	return &FC2DB{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (fc2db *FC2DB) NormalizeMovieID(id string) string {
	return fc2util.ParseNumber(id)
}

func (fc2db *FC2DB) NormalizeMovieKeyword(keyword string) string {
	// 只认 FC2 / FC2-PPV / 纯数字这类 FC2 番号。
	// 其他番号（MIFD-062、SSIS-123…）一律返回空串，引擎据此直接跳过本 provider，
	// 否则每次普通搜索 FC2DB 都会掺一脚。这跟 fc2hub 的做法一致。
	return fc2util.ParseNumber(keyword)
}

func (fc2db *FC2DB) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	// 调用方可能直接传 "FC2-4573362" / "FC2PPV-4573362"，统一成纯数字再拼 URL
	if n := fc2util.ParseNumber(id); n != "" {
		id = n
	}
	return fc2db.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (fc2db *FC2DB) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	segments := strings.Split(strings.Trim(path.Clean(homepage.Path), "/"), "/")
	if len(segments) < 2 || segments[len(segments)-2] != "work" {
		return "", fmt.Errorf("fc2db: invalid work url: %s", rawURL)
	}
	if id := fc2util.ParseNumber(segments[len(segments)-1]); id != "" {
		return id, nil
	}
	return "", fmt.Errorf("fc2db: no valid id in url: %s", rawURL)
}

func (fc2db *FC2DB) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := fc2db.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("FC2-%s", id),
		Provider:      fc2db.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := fc2db.ClonedCollector()

	// Title & Number: [FC2-PPV-4922101] タイトル...
	c.OnXML(`//h1`, func(e *colly.XMLElement) {
		text := strings.TrimSpace(e.Text)
		if ss := titlePattern.FindStringSubmatch(text); len(ss) == 3 {
			info.ID = ss[1]
			info.Number = fmt.Sprintf("FC2-%s", ss[1])
			info.Title = cleanText(ss[2])
			return
		}
		// 没有 [FC2-PPV-xxx] 前缀时，退化为整段文本
		info.Title = cleanText(text)
	})

	// 类型/标签：<a href="/work-tags/ハメ撮り/">ハメ撮り</a>
	c.OnXML(`//a[contains(@href,'/work-tags/')]`, func(e *colly.XMLElement) {
		genre := cleanText(e.Text)
		if genre == "" {
			return
		}
		for _, g := range info.Genres {
			if g == genre {
				return
			}
		}
		info.Genres = append(info.Genres, genre)
	})

	// Cover image.
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if cover := strings.TrimSpace(e.Attr("content")); cover != "" {
			info.CoverURL = resolveImageURL(e.Request.AbsoluteURL(cover))
			info.BigCoverURL = info.CoverURL
			info.ThumbURL = info.CoverURL
		}
	})

	// Actresses: <span>出演女優：</span> <a href="/actress/2982/">name</a>
	c.OnXML(`//span[contains(text(),'出演女優')]/following-sibling::a`, func(e *colly.XMLElement) {
		if name := strings.TrimSpace(e.Text); name != "" {
			info.Actors = append(info.Actors, name)
		}
	})

	// Detail list: 販売日 / 収録時間 / 商品ID / 作品販売者 ...
	c.OnXML(`//dl[contains(@class,'grid')]`, func(e *colly.XMLElement) {
		root, ok := e.DOM.(*html.Node)
		if !ok {
			return
		}
		for _, dt := range htmlquery.Find(root, `.//dt`) {
			dd := htmlquery.FindOne(dt, `following-sibling::dd[1]`)
			if dd == nil {
				continue
			}
			label := strings.TrimSpace(htmlquery.InnerText(dt))
			value := strings.TrimSpace(htmlquery.InnerText(dd))
			switch {
			case strings.HasPrefix(label, "販売日"):
				info.ReleaseDate = parser.ParseDate(value)
			case strings.HasPrefix(label, "収録時間"):
				// 时长由播放端 ffprobe 实际文件得出，元数据里不再填
			case strings.HasPrefix(label, "商品ID"):
				if v := fc2util.ParseNumber(value); v != "" {
					info.ID = v
					info.Number = fmt.Sprintf("FC2-%s", v)
				}
			case strings.HasPrefix(label, "作品販売者"):
				info.Maker = value
			}
		}
	})

	// 预告片不拿（FC2 的 sample mp4 有防盗链，Emby 取流带不上 Referer，播不了），
	// 但接口里那张 1280x720 的横版剧照要收进 preview_images。
	c.OnScraped(func(_ *colly.Response) {
		if info.ID == "" {
			return
		}
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			data := struct {
				Poster string `json:"poster_image_path"`
				Code   int    `json:"code"`
			}{}
			if json.Unmarshal(r.Body, &data) != nil || data.Code != http.StatusOK {
				return
			}
			if strings.HasPrefix(data.Poster, "http") {
				info.PreviewImages = append(info.PreviewImages, data.Poster)
			}
		})
		_ = d.Visit(fmt.Sprintf(sampleAPIURL, info.ID))
	})

	err = c.Visit(info.Homepage)
	return
}

func (fc2db *FC2DB) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	// 纯番号：直接按 ID 取详情。
	// 站内搜索对番号是无效的，只会返回“おすすめ作品”推荐位，
	// 那些跟你要的番号毫无关系，直接查详情才能保证只出一条。
	if id := fc2util.ParseNumber(keyword); id != "" {
		info, e := fc2db.GetMovieInfoByID(id)
		if e != nil || info == nil {
			return nil, nil
		}
		return []*model.MovieSearchResult{info.ToSearchResult()}, nil
	}

	c := fc2db.ClonedCollector()
	seen := make(map[string]bool)
	noResult := false

	// 无结果页会挂“おすすめ作品”，那不是搜索结果，全部丢弃
	c.OnXML(`//h2[contains(text(),'おすすめ')]`, func(*colly.XMLElement) {
		noResult = true
	})

	c.OnXML(`//a[contains(@href,'/work/')]`, func(e *colly.XMLElement) {
		href := e.Request.AbsoluteURL(e.Attr("href"))
		id, err := fc2db.ParseMovieIDFromURL(href)
		if err != nil || id == "" || seen[id] {
			return
		}
		// 卡片标题在 img 的 alt 上，少数卡片没有 alt，退回链接文本
		title := strings.TrimSpace(e.ChildAttr(`.//img`, "alt"))
		if title == "" {
			title = strings.TrimSpace(e.Text)
		}
		if title == "" { // 没标题的结果对搜索没有意义
			return
		}
		seen[id] = true
		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   fmt.Sprintf("FC2-%s", id),
			Title:    cleanText(title),
			Provider: fc2db.Name(),
			Homepage: fmt.Sprintf(movieURL, id),
			CoverURL: resolveImageURL(e.Request.AbsoluteURL(e.ChildAttr(`.//img`, "src"))),
		})
	})

	if err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword))); err != nil {
		return nil, err
	}
	if noResult {
		return nil, nil
	}
	return results, nil
}

func init() {
	provider.Register(Name, New)
}

// resolveImageURL 站点的图片分放在 fc2db.net 和 img.fc2db.net 两个域上，
// 同一路径在错的那个域上直接 404 —— 喂给 Emby 就是破图。
// 这里探一下，返回真正取得到图的 URL。
func resolveImageURL(rawURL string) string {
	if rawURL == "" || !strings.Contains(rawURL, "/wp-content/") {
		return rawURL
	}

	candidates := []string{rawURL}
	if strings.Contains(rawURL, "://fc2db.net/") {
		candidates = []string{
			strings.Replace(rawURL, "://fc2db.net/", "://img.fc2db.net/", 1),
			rawURL,
		}
	}

	client := &http.Client{Timeout: 8 * time.Second}
	for _, cand := range candidates {
		req, err := http.NewRequest(http.MethodGet, cand, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
		req.Header.Set("Range", "bytes=0-0")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent {
			return cand
		}
	}
	return rawURL
}

// cleanText 折叠空白（含全角空格），去掉首尾空白。
// FC2 的标题里常见 “。　※…” 这种全角空格，直接喂给 Emby 会显示成怪异的空隙。
func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\u3000", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "	", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}
