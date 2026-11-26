package javfree

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*JAVFREE)(nil)
	_ provider.MovieSearcher = (*JAVFREE)(nil)
)

const (
	Name     = "JAVFREE"
	Priority = 1000 - 7
)

const (
	baseURL   = "https://javfree.me/"
	movieURL  = "https://javfree.me/ja/%s/" // CloudFlare 301-redirects to the real article page on access
	searchURL = "https://javfree.me/?s=%s"
	// Cover and preview images are hosted on the cf.javfree.me Cloudflare CDN.
	// IMPORTANT: this CDN sits behind a Cloudflare Managed Challenge that blocks
	// headless HTTP clients — metatube-server's fetch receives a 200 response that
	// actually wraps an HTML challenge page, never the real JPEG. A real browser
	// passes the challenge, so the image loads fine when you open the site manually,
	// but the server can NEVER download it. WORKAROUND: resolveCoverFromDMM() borrows
	// the cover art from DMM's pics.dmm.co.jp CDN, which is reachable server-side and
	// returns the real image, so artwork works without switching providers by hand.
	coverURL = "https://cf.javfree.me/cover/500x333/%s.jpg"
)

// Title looks like "[HD][MIDE-980] Movie Title Actress"; it may start with multiple bracketed blocks:
// The first is usually a studio/spec tag (e.g. [HD], [FHD]); the last one is the real ID.
var (
	reLeadingBrackets = regexp.MustCompile(`(?i)^\s*((?:\[[A-Za-z0-9_-]+\]\s*)+)`)
	reBrackets        = regexp.MustCompile(`\[([A-Za-z0-9_-]+)\]`)
	// The ID must contain digits and be >=4 chars to exclude spec tags like HD/FHD/UHD
	// (Go's regexp has no lookahead, so we use two checks: length/charset + must contain digits)
	reNumberLike = regexp.MustCompile(`^[A-Za-z0-9-]{4,}$`)
	reHasDigit   = regexp.MustCompile(`[0-9]`)
	// Release date is shown as "発売日：YYYY/MM/DD" inside a <p>; extract the date part.
	reReleaseDate = regexp.MustCompile(`(\d{4}[/-]\d{1,2}[/-]\d{1,2})`)
	// Skip the metadata info-block lines (品番/監督/メーカー/...) when extracting the synopsis.
	// Labels differ between articles: the cast is "出演者" on some and "出演女優" on others, and
	// the release day is "発売日" on some and "販売日" on others, so both spellings are listed.
	reSummarySkip = regexp.MustCompile(`品番|シリーズ|発売日|販売日|配信開始日|配信日|収録時間|出演者|出演女優|監督|メーカー|レーベル|ジャンル|公開日`)
	// Download links are listed right below the info block on some articles; they are not part
	// of the synopsis and are long enough to pass the minimum length check, so drop them.
	reSummaryURL = regexp.MustCompile(`https?://`)
	// DMM/FANZA content id = series letters + digits, e.g. "mida" + "277".
	reDMMCID = regexp.MustCompile(`^([a-z]+)(\d+)$`)
)

// Split the leading bracketed blocks of the title; returns (candidate IDs, stripped title body)
func splitLeadingBrackets(text string) (candidates []string, body string) {
	m := reLeadingBrackets.FindStringSubmatch(text)
	if len(m) != 2 {
		return nil, strings.TrimSpace(text)
	}
	for _, g := range reBrackets.FindAllStringSubmatch(m[1], -1) {
		if len(g) == 2 {
			candidates = append(candidates, strings.ToUpper(g[1]))
		}
	}
	return candidates, strings.TrimSpace(text[len(m[1]):])
}

// The page's Tags section mixes actor and series/studio tags in one place (e.g. Tags: <a>Actor</a>, <a>Amateur AV</a>, <a>SIRO</a>),
// which cannot be told apart structurally. Heuristic: actor names are almost pure Japanese (kana/kanji) while series/studio tags contain ASCII letters
// (e.g. "Amateur AV" contains AV, "SIRO" is a studio name), so we drop tags containing ASCII letters.
var reASCII = regexp.MustCompile(`[A-Za-z]`)

type JAVFREE struct {
	*fetch.Fetcher
	*scraper.Scraper
}

func New() *JAVFREE {
	return &JAVFREE{
		Fetcher: fetch.Default(&fetch.Config{Referer: baseURL}),
		Scraper: scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese),
	}
}

func (javfree *JAVFREE) NormalizeMovieID(id string) string {
	return strings.ToUpper(strings.TrimSpace(id))
}

func (javfree *JAVFREE) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return javfree.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (javfree *JAVFREE) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.Trim(homepage.Path, "/"), "/")
	if len(parts) < 2 {
		return "", provider.ErrInvalidURL
	}
	// Last segment is the ID, e.g. /119928/ipx-001 -> ipx-001
	return javfree.NormalizeMovieID(parts[len(parts)-1]), nil
}

// dmmProbe is a dedicated fetcher without the retryable client: probing for cover images must
// fail fast, because retrying an unreachable CDN on every scrape would visibly slow things down.
//
// Redirects are deliberately disabled: DMM answers with a 302 pointing at its "now printing"
// placeholder (imgsrc.dmm.com/.../now_printing.jpg) for content ids that do not exist. Following
// that redirect turns a missing cover into a perfectly valid-looking 200 image/jpeg response,
// which is exactly how a placeholder once ended up being served as a real cover.
var dmmProbe = fetch.New(&http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}, &fetch.Config{Referer: baseURL})

// dmmCDNDown remembers that the DMM cover CDN turned out to be unreachable, so later scrapes
// skip it entirely instead of waiting for another round of failed probes. It is only set on a
// network-level failure: a plain miss just means this particular title has no artwork on DMM,
// which must not disable cover borrowing for every other title.
var dmmCDNDown atomic.Bool

// paragraphLines splits a <p> element into lines on <br>. A few articles (DLDSS-522 for one)
// pack the metadata info block and the synopsis into a single <p>, so the two can only be told
// apart line by line rather than paragraph by paragraph.
func paragraphLines(e *colly.XMLElement) []string {
	// OnXML callbacks are served by htmlquery, which stores a *html.Node in DOM.
	node, ok := e.DOM.(*html.Node)
	if !ok {
		return nil
	}
	var sb strings.Builder
	appendNodeText(node, &sb)
	var lines []string
	for _, line := range strings.Split(sb.String(), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// appendNodeText writes the text of n into sb, turning <br> into a newline. Text nodes are
// already unescaped by the parser, so no HTML unescaping is needed here.
func appendNodeText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
		return
	}
	if n.Type == html.ElementNode && strings.EqualFold(n.Data, "br") {
		sb.WriteString("\n")
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		appendNodeText(c, sb)
	}
}

// padDigits left-pads digits with zeros up to the given width.
func padDigits(digits string, width int) string {
	if len(digits) >= width {
		return digits
	}
	return strings.Repeat("0", width-len(digits)) + digits
}

// probeResult is the outcome of probing one candidate image URL.
type probeResult int

const (
	// probeMissing means the CDN answered, but has no such image (DMM 302s to a placeholder).
	probeMissing probeResult = iota
	// probeFound means the URL really serves an image.
	probeFound
	// probeUnreachable means the CDN could not be reached at all (network error, timeout).
	probeUnreachable
)

// probeImage reports whether url really points to a downloadable image. Only the response
// headers are inspected (the body is never read), so probing stays cheap.
func (javfree *JAVFREE) probeImage(url string) probeResult {
	resp, err := dmmProbe.Get(url)
	if err != nil {
		// dmmProbe does not retry, so any error here is a genuine connectivity problem.
		return probeUnreachable
	}
	code, ctype := resp.StatusCode, resp.Header.Get("Content-Type")
	_ = resp.Body.Close()
	if code == http.StatusOK && strings.HasPrefix(ctype, "image/") {
		return probeFound
	}
	// A 3xx is DMM redirecting to its "now printing" placeholder for an unknown content id;
	// redirects are not followed, so it arrives here instead of masquerading as a 200 image.
	return probeMissing
}

// resolveCoverFromDMM returns reachable cover and thumb image URLs on DMM's CDN (pics.dmm.co.jp)
// for the given movie id; both are "" when nothing is reachable. JAVFREE hosts all of its artwork
// on cf.javfree.me, which is unreachable server-side because of a Cloudflare Managed Challenge
// (see the coverURL const), so cover art is borrowed automatically from DMM's cover-stable CDN
// instead of asking the user to switch providers by hand.
func (javfree *JAVFREE) resolveCoverFromDMM(id string) (cover, thumb string) {
	if dmmCDNDown.Load() {
		return "", ""
	}
	// DMM content ids are lowercase and carry no separators: "MIDA-277" -> "mida" + "277".
	base := strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(id))
	m := reDMMCID.FindStringSubmatch(base)
	if m == nil {
		return "", ""
	}
	prefix, digits := m[1], m[2]
	// DMM zero-pads the numeric part of its content ids, but the width differs per series
	// (5 digits for most modern series, 3 for some others). On top of that a few series carry
	// a leading "1" that is part of the content id itself (DLDSS-522 is "1dldss00522", not
	// "dldss00522"), so every width is tried with and without that prefix, most likely first.
	var cids []string
	seen := make(map[string]bool)
	for _, series := range []string{prefix, "1" + prefix} {
		for _, width := range []int{5, 3} {
			if cid := series + padDigits(digits, width); !seen[cid] {
				seen[cid] = true
				cids = append(cids, cid)
			}
		}
	}
	// "pl" is the large package image and "ps" the small one. The large one is roughly an order
	// of magnitude bigger (800x538 vs 147x200) and is the right pick for the cover, because
	// metatube resizes artwork to the configured poster ratio and a small source would only be
	// upscaled. The small one is good enough for the thumb and avoids shipping a large image.
	var hit string
	for _, cid := range cids {
		large := fmt.Sprintf("https://pics.dmm.co.jp/digital/video/%s/%spl.jpg", cid, cid)
		switch javfree.probeImage(large) {
		case probeFound:
			hit, cover = cid, large
		case probeUnreachable:
			dmmCDNDown.Store(true)
			return "", ""
		}
		if hit != "" {
			break
		}
	}
	if hit == "" {
		// JAVFREE carries titles DMM does not have; that is a miss, not a broken CDN.
		return "", ""
	}
	thumb = fmt.Sprintf("https://pics.dmm.co.jp/digital/video/%s/%sps.jpg", hit, hit)
	if javfree.probeImage(thumb) != probeFound {
		thumb = cover
	}
	return cover, thumb
}

func (javfree *JAVFREE) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := javfree.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        id,
		Provider:      javfree.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
		// NOTE: these default to cf.javfree.me, which is unreachable server-side
		// (Cloudflare Managed Challenge, see the coverURL const). They are replaced
		// by resolveCoverFromDMM() with a reachable DMM CDN image when one is found,
		// so artwork works without switching providers by hand.
		CoverURL: fmt.Sprintf(coverURL, strings.ToUpper(id)),
		ThumbURL: fmt.Sprintf(coverURL, strings.ToUpper(id)),
	}

	c := javfree.ClonedCollector()

	// Title + ID
	c.OnXML(`//h1[@class="entry-title"]`, func(e *colly.XMLElement) {
		text := strings.TrimSpace(e.Text)
		candidates, body := splitLeadingBrackets(text)
		info.Title = body
		// Take the last ID-like prefix block (e.g. [HD][MIDE-980] -> MIDE-980)
		for i := len(candidates) - 1; i >= 0; i-- {
			if reNumberLike.MatchString(candidates[i]) && reHasDigit.MatchString(candidates[i]) {
				info.Number = candidates[i]
				break
			}
		}
	})

	// Actors: non-ASCII (Japanese) tags in rel="tag"
	c.OnXML(`//a[@rel="tag"]`, func(e *colly.XMLElement) {
		if tag := strings.TrimSpace(e.Text); tag != "" && !reASCII.MatchString(tag) {
			info.Actors = append(info.Actors, tag)
		}
	})

	// Preview images: entry-content is filled by JS dynamically; best-effort only
	c.OnXML(`//div[@class="entry-content"]//img`, func(e *colly.XMLElement) {
		if src := strings.TrimSpace(e.Attr("src")); src != "" {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(src))
		}
	})

	// Summary: the article body's narrative <p> blocks inside entry-content form the
	// synopsis. The real synopsis often spans several <p> blocks, so collect them all
	// (joined by newline) instead of keeping only the first. The metadata info-block
	// <p> (品番/監督/メーカー/...) is skipped as it is not part of the synopsis.
	c.OnXML(`//div[@class="entry-content"]//p`, func(e *colly.XMLElement) {
		text := strings.TrimSpace(e.Text)
		lines := []string{text}
		if reSummarySkip.MatchString(text) {
			// Some articles keep the info block and the synopsis in one and the same <p>,
			// separated by <br> (DLDSS-522 does), so discarding the paragraph wholesale would
			// throw the synopsis away together with the metadata. Split it and keep the
			// narrative lines only.
			lines = paragraphLines(e)
		}
		for _, line := range lines {
			if len([]rune(line)) < 20 || reSummarySkip.MatchString(line) || reSummaryURL.MatchString(line) {
				continue
			}
			if info.Summary != "" {
				info.Summary += "\n"
			}
			info.Summary += line
		}
	})

	// Release date: written as "発売日：YYYY/MM/DD" or "配信開始日：YYYY-MM-DD" inside a <p>.
	// The site uses different labels across titles, so match any of the common release-day labels.
	c.OnXML(`//p[contains(., '発売日') or contains(., '配信開始日') or contains(., '販売日') or contains(., '公開日')]`, func(e *colly.XMLElement) {
		if m := reReleaseDate.FindStringSubmatch(e.Text); m != nil {
			info.ReleaseDate = parser.ParseDate(m[1])
		}
	})

	if err = c.Visit(info.Homepage); err != nil {
		return
	}
	// JAVFREE's own artwork CDN is unreachable server-side (see the coverURL const), so borrow
	// cover art from DMM's CDN automatically whenever an image is actually reachable there,
	// instead of requiring the user to switch providers by hand.
	if cover, thumb := javfree.resolveCoverFromDMM(info.Number); cover != "" {
		info.CoverURL = cover
		if thumb != "" {
			info.ThumbURL = thumb
		}
	}
	return
}

func (javfree *JAVFREE) NormalizeMovieKeyword(keyword string) string {
	return strings.ToUpper(strings.TrimSpace(keyword))
}

func (javfree *JAVFREE) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	c := javfree.ClonedCollector()
	c.OnXML(`//article[contains(@class,"hentry")]`, func(e *colly.XMLElement) {
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//h2[@class="entry-title"]/a`, "href"))
		id, _ := javfree.ParseMovieIDFromURL(homepage)
		title := strings.TrimSpace(e.ChildText(`.//h2[@class="entry-title"]/a`))
		thumb := e.Request.AbsoluteURL(e.ChildAttr(`.//div[@class="thumbnail-wrap"]/img`, "src"))
		cover := ""
		if id != "" {
			cover = fmt.Sprintf(coverURL, strings.ToUpper(id))
		}
		results = append(results, &model.MovieSearchResult{
			ID:       id,
			Number:   id,
			Title:    title,
			Provider: javfree.Name(),
			Homepage: homepage,
			ThumbURL: thumb,
			CoverURL: cover,
		})
	})
	err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword)))
	return
}

func init() {
	provider.Register(Name, New)
}
