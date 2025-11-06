package dlsite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/xmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"
	"gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*DLsite)(nil)
	_ provider.MovieSearcher = (*DLsite)(nil)
)

const (
	Name     = "DLsite"
	Priority = 1000 - 3
)

const (
	baseURL          = "https://www.dlsite.com/"
	maniaxWorkURL    = "https://www.dlsite.com/maniax/work/=/product_id/%s.html"
	proWorkURL       = "https://www.dlsite.com/pro/work/=/product_id/%s.html"
	unifiedSearchURL = "https://www.dlsite.com/maniax/fsr/=/keyword/%s/work_type_category[0]/movie/"
)

/* =========================
   Precompiled regex
   ========================= */

var (
	rjRe            = regexp.MustCompile(`(?i)^RJ\d+$`)
	vjRe            = regexp.MustCompile(`(?i)^VJ\d+$`)
	productPathRe   = regexp.MustCompile(`(?i)^(RJ|VJ)\d+\.html$`)
	spaceCollapseRe = regexp.MustCompile(`\s+`)
	dateJPRe        = regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日`)
	parenRe         = regexp.MustCompile(`[()（）][^()（）]*[()（）]`)
)

/* =========================
   Shared HTTP resources (reduce total connections)
   ========================= */

// sharedTransport is reused across all requests/collectors to minimize new connections.
var sharedTransport = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          128,
	MaxIdleConnsPerHost:   64,
	MaxConnsPerHost:       16, // limit concurrent connections per host
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   20 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// sharedHTTPClient wraps the shared transport for direct JSON calls.
var sharedHTTPClient = &http.Client{
	Transport: sharedTransport,
	Timeout:   30 * time.Second,
}

/* =========================
   Types / Constructor
   ========================= */

type DLsite struct {
	*scraper.Scraper
}

func New() *DLsite {
	return &DLsite{
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese),
	}
}

/* =========================
   Utilities (merged & generic)
   ========================= */

// normalizeSpaces replaces full-width spaces, collapses whitespace and trims.
func normalizeSpaces(s string) string {
	s = strings.ReplaceAll(s, "　", " ")
	return strings.TrimSpace(spaceCollapseRe.ReplaceAllString(s, " "))
}

// titleClean normalizes title text.
func titleClean(s string) string { return normalizeSpaces(s) }

// textCleanSoft keeps paragraph breaks but collapses excessive blank lines and spaces.
func textCleanSoft(s string) string {
	s = strings.ReplaceAll(s, "　", " ")
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")
	return spaceCollapseRe.ReplaceAllString(s, " ")
}

// normalizeName collapses spaces and trims NBSP/full-width spaces.
func normalizeName(s string) string {
	s = strings.ReplaceAll(s, "\u00A0", " ")
	s = strings.ReplaceAll(s, "　", " ")
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// absURL builds an absolute URL from possibly relative input.
func absURL(u string, e *colly.XMLElement) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "//") {
		return "https:" + u
	}
	if strings.HasPrefix(u, "/") && e != nil {
		return e.Request.AbsoluteURL(u)
	}
	return u
}

// textWithBRXML extracts text from a xmlquery.Node and converts <br> to '\n'.
func textWithBRXML(n *xmlquery.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var walk func(*xmlquery.Node)
	walk = func(x *xmlquery.Node) {
		if x == nil {
			return
		}
		if x.Type == xmlquery.TextNode {
			b.WriteString(x.Data)
		}
		if strings.EqualFold(x.Data, "br") {
			b.WriteString("\n")
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

// textWithBRHTML extracts text from a goquery.Selection and converts <br> to '\n'.
func textWithBRHTML(sel *goquery.Selection) string {
	var b strings.Builder
	sel.Contents().Each(func(_ int, s *goquery.Selection) {
		if n := s.Get(0); n != nil && strings.EqualFold(n.Data, "br") {
			b.WriteString("\n")
			return
		}
		if t := s.Text(); t != "" {
			b.WriteString(t)
		}
	})
	return b.String()
}

// drainAndClose ensures body is drained so that the connection can be reused.
func drainAndClose(rc io.ReadCloser) {
	if rc == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(rc, 512))
	_ = rc.Close()
}

// extractOutlineCell returns normalized text from a <td>, preferring the first <a>.
func extractOutlineCell(e *colly.XMLElement) string {
	val := strings.TrimSpace(e.ChildText(`.//a`))
	if val == "" {
		val = strings.TrimSpace(e.Text)
	}
	return normalizeSpaces(val)
}

// pickJPG returns a JPG url from src/srcset, converting .webp→.jpg when possible.
func pickJPG(src, srcset string, e *colly.XMLElement) string {
	if s := absURL(src, e); s != "" {
		ls := strings.ToLower(s)
		if strings.HasSuffix(ls, ".jpg") {
			return s
		}
		if strings.HasSuffix(ls, ".webp") {
			return strings.TrimSuffix(s, path.Ext(s)) + ".jpg"
		}
	}
	for _, part := range strings.Split(srcset, ",") {
		p := strings.TrimSpace(strings.Split(part, " ")[0])
		if strings.HasSuffix(strings.ToLower(p), ".jpg") {
			return absURL(p, e)
		}
	}
	return ""
}

// addPreviewImage appends a preview image if unique; .webp will be mapped to .jpg.
func addPreviewImage(info *model.MovieInfo, u string, e *colly.XMLElement) {
	u = absURL(strings.TrimSpace(u), e)
	if u == "" {
		return
	}
	if strings.HasSuffix(strings.ToLower(u), ".webp") {
		u = strings.TrimSuffix(u, path.Ext(u)) + ".jpg"
	}
	for _, existed := range info.PreviewImages {
		if existed == u {
			return
		}
	}
	info.PreviewImages = append(info.PreviewImages, u)
}

// cleanPerson removes content inside parentheses, leading symbols and trims.
func cleanPerson(s string) string {
	s = strings.ReplaceAll(s, "\u00A0", " ")
	s = strings.ReplaceAll(s, "　", " ")
	s = parenRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.TrimLeft(s, "＞>・-— ")
	return strings.TrimSpace(s)
}

// setDirectorOnce sets Director only if it's empty (with person cleanup).
func setDirectorOnce(info *model.MovieInfo, name string) {
	if info.Director != "" {
		return
	}
	name = cleanPerson(name)
	if name != "" {
		info.Director = name
	}
}

// addActorUnique normalizes and adds an actor if unique.
func addActorUnique(info *model.MovieInfo, name string) {
	name = normalizeName(name)
	if name == "" {
		return
	}
	for _, a := range info.Actors {
		if a == name {
			return
		}
	}
	info.Actors = append(info.Actors, name)
}

// sanitizeKeyword replaces unusual symbols with spaces while keeping letters, numbers and a few safe chars.
// This helps cases like "祓魔○女シャルロット" where '○' should not break search.
func sanitizeKeyword(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		if unicode.IsSpace(r) {
			return ' '
		}
		// keep a few separators that often appear in titles
		switch r {
		case '-', '_':
			return r
		default:
			return ' ' // unusual symbols (e.g., '○') -> space
		}
	}, s)
	// collapse multiple spaces
	return strings.TrimSpace(spaceCollapseRe.ReplaceAllString(s, " "))
}

/* =========================
   AJAX rating fetch
   ========================= */

// fetchRateAverage2dp fetches score with two decimals. Null score returns 0.
func (d *DLsite) fetchRateAverage2dp(homepageURL, id string) (float64, error) {
	site := "maniax"
	if strings.Contains(homepageURL, "/pro/") {
		site = "pro"
	}
	ajaxURL := fmt.Sprintf("https://www.dlsite.com/%s/product/info/ajax?product_id=%s", site, id)

	req, err := http.NewRequest(http.MethodGet, ajaxURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", homepageURL)
	req.Header.Set("Accept", "application/json")

	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer drainAndClose(resp.Body)

	// Parse only rate_average_2dp
	var m map[string]struct {
		RateAverage2dp *float64 `json:"rate_average_2dp"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return 0, err
	}
	item, ok := m[id]
	if !ok || item.RateAverage2dp == nil {
		return 0, nil
	}
	return *item.RateAverage2dp, nil
}

/* =========================
   ID & URL helpers
   ========================= */

func (d *DLsite) NormalizeMovieID(id string) string {
	return strings.ToUpper(strings.TrimSpace(id))
}

func (d *DLsite) ParseMovieIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	base := path.Base(u.Path)
	if productPathRe.MatchString(base) {
		id := strings.TrimSuffix(base, path.Ext(base))
		return d.NormalizeMovieID(id), nil
	}
	for _, p := range strings.Split(u.Path, "/") {
		if rjRe.MatchString(p) || vjRe.MatchString(p) {
			return d.NormalizeMovieID(p), nil
		}
	}
	return "", fmt.Errorf("cannot parse dlsite id from url: %s", rawURL)
}

/* =========================
   Detail: Get by ID / URL
   ========================= */

func (d *DLsite) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	id = d.NormalizeMovieID(id)
	if rjRe.MatchString(id) {
		if info, err = d.GetMovieInfoByURL(fmt.Sprintf(maniaxWorkURL, id)); err == nil && info != nil && info.IsValid() {
			return info, nil
		}
		return d.GetMovieInfoByURL(fmt.Sprintf(proWorkURL, id))
	}
	if vjRe.MatchString(id) {
		if info, err = d.GetMovieInfoByURL(fmt.Sprintf(proWorkURL, id)); err == nil && info != nil && info.IsValid() {
			return info, nil
		}
		return d.GetMovieInfoByURL(fmt.Sprintf(maniaxWorkURL, id))
	}
	if info, err = d.GetMovieInfoByURL(fmt.Sprintf(maniaxWorkURL, id)); err == nil && info != nil && info.IsValid() {
		return info, nil
	}
	return d.GetMovieInfoByURL(fmt.Sprintf(proWorkURL, id))
}

func (d *DLsite) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := d.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}
	info = &model.MovieInfo{
		ID:            id,
		Provider:      d.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := d.ClonedCollector()
	// Reuse shared transport to minimize new connections.
	c.WithTransport(sharedTransport)

	// ===== Title =====
	c.OnXML(`//h1[@id="work_name"]`, func(e *colly.XMLElement) {
		info.Title = titleClean(e.Text)
	})

	// ===== Summary (keep <br> as newline) =====
	c.OnXML(`//div[@itemprop="description" and contains(@class, "work_parts_container")]`, func(e *colly.XMLElement) {
		var typeTextParas, areaParas []string

		switch dom := e.DOM.(type) {
		case *xmlquery.Node:
			for _, node := range xmlquery.Find(dom, `.//div[contains(@class,"work_parts_multitype_item") and contains(@class,"type_text")]`) {
				ps := xmlquery.Find(node, `.//p`)
				if len(ps) > 0 {
					for _, p := range ps {
						if raw := strings.TrimSpace(textWithBRXML(p)); raw != "" {
							typeTextParas = append(typeTextParas, raw)
						}
					}
				} else {
					if raw := strings.TrimSpace(textWithBRXML(node)); raw != "" {
						typeTextParas = append(typeTextParas, raw)
					}
				}
			}
			for _, p := range xmlquery.Find(dom, `.//div[contains(@class,"work_parts_area")]//p`) {
				if raw := strings.TrimSpace(textWithBRXML(p)); raw != "" {
					areaParas = append(areaParas, raw)
				}
			}
		case *html.Node:
			sel := goquery.NewDocumentFromNode(dom).Selection
			sel.Find(`div.work_parts_multitype_item.type_text`).Each(func(_ int, s *goquery.Selection) {
				if s.Find("p").Length() > 0 {
					s.Find("p").Each(func(_ int, p *goquery.Selection) {
						if raw := strings.TrimSpace(textWithBRHTML(p)); raw != "" {
							typeTextParas = append(typeTextParas, raw)
						}
					})
				} else {
					if raw := strings.TrimSpace(textWithBRHTML(s)); raw != "" {
						typeTextParas = append(typeTextParas, raw)
					}
				}
			})
			sel.Find(`div.work_parts_area p`).Each(func(_ int, p *goquery.Selection) {
				if raw := strings.TrimSpace(textWithBRHTML(p)); raw != "" {
					areaParas = append(areaParas, raw)
				}
			})
		default:
			if raw := strings.TrimSpace(e.Text); raw != "" {
				typeTextParas = []string{raw}
			}
		}

		if len(typeTextParas) > 0 || len(areaParas) > 0 {
			all := append([]string{}, typeTextParas...)
			all = append(all, areaParas...)
			joined := textCleanSoft(strings.Join(all, "\n\n"))
			if len(joined) > len(info.Summary) {
				info.Summary = joined
			}
		}
	})

	// ===== Director =====
	// Priority: シナリオ > その他 > イラスト

	// ① シナリオ
	c.OnXML(`//table[@id='work_outline']
          //tr[.//th[contains(normalize-space(.),'シナリオ')]]
          //td`, func(e *colly.XMLElement) {
		name := strings.TrimSpace(e.ChildText(`.//a[1]`))
		if name == "" {
			name = strings.TrimSpace(e.Text)
		}
		setDirectorOnce(info, name)
	})

	// ② その他
	c.OnXML(`//table[@id='work_outline']
          //tr[.//th[contains(normalize-space(.),'その他')]]
          //td`, func(e *colly.XMLElement) {
		if info.Director != "" {
			return
		}
		name := strings.TrimSpace(e.ChildText(`.//a[1]`))
		if name == "" {
			name = strings.TrimSpace(e.Text)
		}
		setDirectorOnce(info, name)
	})

	// ③ イラスト
	c.OnXML(`//table[@id='work_outline']
          //tr[.//th[contains(normalize-space(.),'イラスト')]]
          //td`, func(e *colly.XMLElement) {
		if info.Director != "" {
			return
		}
		name := strings.TrimSpace(e.ChildText(`.//a[1]`))
		if name == "" {
			name = strings.TrimSpace(e.Text)
		}
		setDirectorOnce(info, name)
	})

	// ===== Actors (first matched row among 声優/出演者/キャスト) =====
	c.OnXML(`(
            //*[@id='work_right']//table[@id='work_outline']
            //tr[.//th[contains(normalize-space(.),'声優')
                   or contains(normalize-space(.),'出演者')
                   or contains(normalize-space(.),'キャスト')]]
          )[1]//td//a`, func(e *colly.XMLElement) {
		addActorUnique(info, e.Text)
	})

	// ===== Cover / Big Cover =====
	c.OnXML(`//*[@id="work_left"]//div[contains(@class,'work_slider_container')]
          //li[contains(@class,'slider_item') and contains(@class,'active')]//img`, func(e *colly.XMLElement) {
		if big := pickJPG(e.Attr("src"), e.Attr("srcset"), e); big != "" {
			info.BigThumbURL = big
			info.CoverURL = big
			info.BigCoverURL = big
		}
	})

	// Thumb (240x240) — first data-thumb of product-slider-data
	c.OnXML(`//*[@id='work_left']//div[contains(@class,'product-slider-data')]/div[1]`, func(e *colly.XMLElement) {
		u := strings.TrimSpace(e.Attr("data-thumb"))
		if u == "" {
			return
		}
		info.ThumbURL = absURL(u, e)
	})

	// ===== Preview Images =====
	c.OnXML(`//*[@id='work_left']//div[contains(@class,'product-slider-data')]//div[@data-src]`, func(e *colly.XMLElement) {
		if src := strings.TrimSpace(e.Attr("data-src")); src != "" {
			addPreviewImage(info, src, e)
			return
		}
		if thumb := strings.TrimSpace(e.Attr("data-thumb")); thumb != "" {
			addPreviewImage(info, thumb, e)
		}
	})

	// ===== Maker =====
	c.OnXML(`//table[@id="work_maker"]
          //tr[.//th[contains(normalize-space(.),"ブランド名")
                    or contains(normalize-space(.),"サークル名")]]
          //td`, func(e *colly.XMLElement) {
		if info.Maker != "" {
			return
		}
		if maker := extractOutlineCell(e); maker != "" {
			info.Maker = maker
		}
	})

	// ===== Series =====
	c.OnXML(`//table[@id="work_outline"]
          //tr[.//th[contains(normalize-space(.),"シリーズ")]]
          //td`, func(e *colly.XMLElement) {
		if info.Series != "" {
			return
		}
		if series := extractOutlineCell(e); series != "" {
			info.Series = series
		}
	})

	// ===== Genres (ジャンル) =====
	c.OnXML(`//table[@id="work_outline"]
          //tr[.//th[contains(normalize-space(.),"ジャンル")]]
          //td//div[contains(@class,"main_genre")]//a`, func(e *colly.XMLElement) {
		genre := strings.TrimSpace(e.Text)
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

	// ===== Release Date (販売日) =====
	c.OnXML(`//table[@id="work_outline"]
          //tr[.//th[contains(normalize-space(.),"販売日")]]
          //td`, func(e *colly.XMLElement) {
		raw := strings.TrimSpace(e.ChildText(`.//a`))
		if raw == "" {
			raw = strings.TrimSpace(e.Text)
		}
		if m := dateJPRe.FindStringSubmatch(raw); len(m) == 4 {
			year, _ := strconv.Atoi(m[1])
			month, _ := strconv.Atoi(m[2])
			day, _ := strconv.Atoi(m[3])
			t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			info.ReleaseDate = datatypes.Date(t)
		}
	})

	// ===== Label (作品形式, first item) =====
	c.OnXML(`//table[@id="work_outline"]
          //tr[.//th[contains(normalize-space(.),'作品形式')]]
          //td//div[@id='category_type']//a[1]`, func(e *colly.XMLElement) {
		// Prefer span@title, fallback to text.
		label := strings.TrimSpace(e.ChildAttr(".//span[1]", "title"))
		if label == "" {
			label = strings.TrimSpace(e.Text)
		}
		if label != "" {
			info.Label = label
		}
	})

	// ===== Finalize & Rating =====
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" {
			info.CoverURL = info.ThumbURL
		}
		if info.Number == "" {
			info.Number = info.ID
		}
		if s, err := d.fetchRateAverage2dp(info.Homepage, info.ID); err == nil && s > 0 {
			info.Score = s
		}
	})

	// Fallbacks (idempotent)
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" {
			info.CoverURL = info.ThumbURL
		}
		if info.Number == "" {
			info.Number = info.ID
		}
	})

	err = c.Visit(info.Homepage)
	return
}

/* =========================
   Search
   ========================= */

func (d *DLsite) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	cleaned := sanitizeKeyword(keyword)
	// replace spaces with '+', safe for URL
	return strings.ReplaceAll(cleaned, " ", "+")
}

func (d *DLsite) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	kw := d.NormalizeMovieKeyword(keyword)
	if kw == "" {
		return nil, nil
	}

	var ids []string
	seen := make(map[string]struct{})

	// Use shared transport for search collector as well.
	collect := func(searchFmt string) error {
		c := d.ClonedCollector()
		c.WithTransport(sharedTransport)

		c.OnXML(`//div[@id="search_result_list"]//li//a[contains(@href,"/work/=/product_id/")]`, func(e *colly.XMLElement) {
			href := e.Request.AbsoluteURL(e.Attr("href"))
			u, _ := url.Parse(href)
			base := path.Base(u.Path)
			id := strings.TrimSuffix(base, path.Ext(base))
			id = d.NormalizeMovieID(id)
			if !rjRe.MatchString(id) && !vjRe.MatchString(id) {
				return
			}
			if _, ok := seen[id]; ok {
				return
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		})

		return c.Visit(fmt.Sprintf(searchFmt, url.PathEscape(kw)))
	}

	_ = collect(unifiedSearchURL)

	// Limit subsequent detail fetches to reduce connection pressure.
	const limit = 5
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for i := 0; i < len(ids) && i < limit; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if info, _ := d.GetMovieInfoByID(ids[i]); info != nil && info.IsValid() {
				mu.Lock()
				results = append(results, info.ToSearchResult())
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return
}

/* =========================
   Registration
   ========================= */

func init() {
	provider.Register(Name, New)
}
