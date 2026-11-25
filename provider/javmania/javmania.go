// Package javmania provides English metadata for uncensored JAV studios
// (Caribbeancom, 1Pondo, PacoPacoMama, HEYZO, ...) and censored releases, plus
// performer profiles, from https://www.jav-mania.net/. Japanese performer names
// are resolved through the companion catalogue https://avhinban.com/.
package javmania

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*JAVMania)(nil)
	_ provider.MovieSearcher = (*JAVMania)(nil)
	_ provider.ActorProvider = (*JAVMania)(nil)
	_ provider.ActorSearcher = (*JAVMania)(nil)
)

const (
	Name = "JAVMANIA"
	// Below the official studio providers: users who want English metadata
	// raise it with MT_MOVIE_PROVIDER_JAVMANIA__PRIORITY.
	Priority = 1000 - 6
)

const (
	baseURL         = "https://www.jav-mania.net/"
	movieURL        = "https://www.jav-mania.net/movies/%s/%s/"
	javURL          = "https://www.jav-mania.net/jav/%s/"
	actorURL        = "https://www.jav-mania.net/actress/%s/"
	avhinbanSearch  = "https://avhinban.com/search/?q=%s"
	avhinbanBaseURL = "https://avhinban.com"
)

// site is one uncensored studio on JAV Mania.
type site struct {
	slug    string   // path segment on jav-mania.net and movie-ID prefix
	aliases []string // spellings seen in file names
	id      *regexp.Regexp
	format  func([]string) string // canonical content ID from id submatches
	number  func(string) string   // movie number shown to users
	valid   *regexp.Regexp        // canonical content ID
}

var (
	d6     = `(\d{6})`
	sep    = `[-_ .]`
	asIs   = func(m []string) string { return m[1] }
	dash   = func(m []string) string { return m[1] + "-" + m[2] }
	under  = func(m []string) string { return m[1] + "_" + m[2] }
	plain  = func(id string) string { return id }
	prefix = func(p string) func(string) string { return func(id string) string { return p + id } }
)

var sites = []site{
	{"caribbeancom", []string{"caribbeancom", "carib"}, regexp.MustCompile(d6 + sep + `(\d{3})`), dash, plain, regexp.MustCompile(`^\d{6}-\d{3}$`)},
	{"caribbeanpremium", []string{"caribbeancompr", "caribbeanpremium", "carib-pr", "carib_pr", "caribpr"}, regexp.MustCompile(d6 + sep + `(\d{3})`), under, plain, regexp.MustCompile(`^\d{6}_\d{3}$`)},
	{"1pondo", []string{"1pondo", "1pon"}, regexp.MustCompile(d6 + sep + `(\d{3})`), under, plain, regexp.MustCompile(`^\d{6}_\d{3}$`)},
	{"pacopacomama", []string{"pacopacomama", "paco"}, regexp.MustCompile(d6 + sep + `(\d{3})`), under, plain, regexp.MustCompile(`^\d{6}_\d{3}$`)},
	{"10musume", []string{"10musume", "10mu"}, regexp.MustCompile(d6 + sep + `(\d{2})`), under, plain, regexp.MustCompile(`^\d{6}_\d{2}$`)},
	{"muramura", []string{"muramura"}, regexp.MustCompile(d6 + sep + `(\d{3,4})`), under, plain, regexp.MustCompile(`^\d{6}_\d{3,4}$`)},
	{"heyzo", []string{"heyzo"}, regexp.MustCompile(`(?:hd` + sep + `?)?(\d{4})`), asIs, prefix("HEYZO-"), regexp.MustCompile(`^\d{4}$`)},
	{"kin8", []string{"kin8tengoku", "kin8"}, regexp.MustCompile(`(\d{4})`), asIs, prefix("KIN8-"), regexp.MustCompile(`^\d{4}$`)},
	{"nyoshin", []string{"nyoshin"}, regexp.MustCompile(`(n\d{3,5})`), asIs, plain, regexp.MustCompile(`^n\d{3,5}$`)},
	{"hgmo", []string{"heydouga", "hgmo"}, regexp.MustCompile(`(\d{4})`), asIs, prefix("HGMO-"), regexp.MustCompile(`^\d{4}$`)},
	{"h4610", []string{"h4610"}, regexp.MustCompile(`([a-z]{2,6}\d{5,6})`), asIs, plain, regexp.MustCompile(`^[a-z]{2,6}\d{5,6}$`)},
	{"c0930", []string{"c0930"}, regexp.MustCompile(`([a-z]{2,6}\d{5,6})`), asIs, plain, regexp.MustCompile(`^[a-z]{2,6}\d{5,6}$`)},
	{"h0930", []string{"h0930"}, regexp.MustCompile(`([a-z]{2,6}\d{5,6})`), asIs, plain, regexp.MustCompile(`^[a-z]{2,6}\d{5,6}$`)},
	{"pikkur", []string{"pikkur"}, regexp.MustCompile(`(\d{5})`), asIs, prefix("PIKKUR-"), regexp.MustCompile(`^\d{5}$`)},
	{"av9898", []string{"av9898"}, regexp.MustCompile(`(\d{4})`), asIs, prefix("AV9898-"), regexp.MustCompile(`^\d{4}$`)},
}

// aliasOrder lists every studio alias, longest first, so "carib-pr" wins over "carib".
var aliasOrder = func() (out []struct {
	alias string
	site  site
},
) {
	for _, st := range sites {
		for _, a := range st.aliases {
			out = append(out, struct {
				alias string
				site  site
			}{a, st})
		}
	}
	slices.SortStableFunc(out, func(a, b struct {
		alias string
		site  site
	},
	) int {
		return len(b.alias) - len(a.alias)
	})
	return
}()

// Bare IDs without a studio name and the studios that use that shape.
var bareIDs = []struct {
	re    *regexp.Regexp
	slugs []string
}{
	{regexp.MustCompile(`(?:^|\D)` + d6 + `-(\d{3})(?:\D|$)`), []string{"caribbeancom"}},
	{regexp.MustCompile(`(?:^|\D)` + d6 + `_(\d{3})(?:\D|$)`), []string{"1pondo", "caribbeanpremium", "pacopacomama"}},
	{regexp.MustCompile(`(?:^|\D)` + d6 + `_(\d{2})(?:\D|$)`), []string{"10musume"}},
}

var (
	allowedHosts = []string{"www.jav-mania.net", "jav-mania.net"}
	// Censored codes: SSIS-001, ssis 001, ssis00001 (zero padded).
	censoredRe = regexp.MustCompile(`(?:^|[^a-z0-9])([a-z]{2,6})(?:[-_ .](\d{2,5})|(0\d{2,4}))(?:\D|$)`)
	notCodes   = []string{"hd", "fhd", "uhd", "mp", "avi", "mkv", "wmv", "part", "cd", "disc", "dvd", "img", "dsc", "vid", "mov", "clip", "scene", "fc", "ppv"}
	javCodeRe  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	actorIDRe  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	nonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)
	extRe      = regexp.MustCompile(`\.(?:mp4|mkv|avi|wmv|mov|m4v|ts|flv|webm|rmvb|iso)$`)
	trailerRe  = regexp.MustCompile(`(\s*\([^()]*\))+$`)
	pagePath   = regexp.MustCompile(`^/(?:[a-z]{2}(?:-[a-z]{2,4})?/)?(?:movies/([a-z0-9]+)/([A-Za-z0-9_-]+)|jav/([A-Za-z0-9_-]+))/?$`)
	actorPath  = regexp.MustCompile(`^/(?:[a-z]{2}(?:-[a-z]{2,4})?/)?actress/([a-z0-9-]+)/?$`)
	profileRe  = struct{ height, cup, bwh, born *regexp.Regexp }{
		regexp.MustCompile(`(\d{3})\s*cm`),
		regexp.MustCompile(`\b([A-Z]{1,2})-cup\b`),
		regexp.MustCompile(`B(\d{2,3})\s*W(\d{2,3})\s*H(\d{2,3})`),
		regexp.MustCompile(`Born\s+(\d{4}-\d{2}-\d{2})`),
	}
)

type JAVMania struct {
	*scraper.Scraper
}

func New() *JAVMania {
	return &JAVMania{scraper.NewDefaultScraper(Name, baseURL, Priority, language.English)}
}

// ---------- movie IDs ----------
//
// Movie IDs are "<studio slug>-<content id>" for uncensored titles
// (e.g. "pacopacomama-101416_183", "caribbeancom-112426-001") and
// "jav-<code>" for censored titles (e.g. "jav-ssis-001").

func siteBySlug(slug string) (site, bool) {
	for _, s := range sites {
		if s.slug == slug {
			return s, true
		}
	}
	return site{}, false
}

// splitID validates a movie ID and splits it into studio slug and content ID.
func splitID(id string) (slug, rest string, ok bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	if code, found := strings.CutPrefix(id, "jav-"); found {
		return "jav", code, javCodeRe.MatchString(code)
	}
	for _, s := range sites {
		if cid, found := strings.CutPrefix(id, s.slug+"-"); found {
			return s.slug, cid, s.valid.MatchString(cid)
		}
	}
	return "", "", false
}

func (jm *JAVMania) NormalizeMovieID(id string) string {
	if slug, rest, ok := splitID(id); ok {
		return slug + "-" + rest
	}
	if cands := movieCandidates(id); len(cands) == 1 {
		return cands[0]
	}
	return ""
}

func isAllowedHost(u *url.URL) bool {
	return u.Scheme == "https" && slices.Contains(allowedHosts, strings.ToLower(u.Hostname()))
}

func (jm *JAVMania) ParseMovieIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	m := pagePath.FindStringSubmatch(path.Clean(u.Path) + "/")
	if m == nil || !isAllowedHost(u) {
		return "", fmt.Errorf("javmania: not a work page: %s", rawURL)
	}
	id := "jav-" + strings.ToLower(m[3])
	if m[3] == "" {
		if _, known := siteBySlug(m[1]); !known {
			return "", fmt.Errorf("javmania: unknown studio in %s", rawURL)
		}
		id = m[1] + "-" + m[2]
	}
	if _, _, ok := splitID(id); !ok {
		return "", fmt.Errorf("javmania: unsupported work page: %s", rawURL)
	}
	return strings.ToLower(id), nil
}

func movieURLFor(id string) (string, error) {
	slug, rest, ok := splitID(id)
	if !ok {
		return "", fmt.Errorf("javmania: invalid movie id: %s", id)
	}
	if slug == "jav" {
		return fmt.Sprintf(javURL, rest), nil
	}
	return fmt.Sprintf(movieURL, slug, rest), nil
}

// movieCandidates turns a keyword or file name into candidate movie IDs.
func movieCandidates(keyword string) []string {
	if slug, rest, ok := splitID(keyword); ok {
		return []string{slug + "-" + rest} // already a movie ID
	}
	s := extRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(keyword)), "")
	for _, h := range aliasOrder {
		at := aliasIndex(s, h.alias)
		if at < 0 {
			continue
		}
		// The ID follows the alias ("paco 101416_183") or, when nothing
		// follows the alias, precedes it ("101416_183-paco").
		if m := adjacentMatch(h.site.id, s[at+len(h.alias):], true); m != nil {
			return []string{h.site.slug + "-" + h.site.format(m)}
		}
		if !strings.ContainsAny(s[at+len(h.alias):], "0123456789") {
			if m := adjacentMatch(h.site.id, s[:at], false); m != nil {
				return []string{h.site.slug + "-" + h.site.format(m)}
			}
		}
		return nil
	}
	for _, b := range bareIDs {
		if m := b.re.FindStringSubmatch(s); m != nil {
			ids := make([]string, 0, len(b.slugs))
			for _, slug := range b.slugs {
				st, _ := siteBySlug(slug)
				ids = append(ids, slug+"-"+st.format([]string{m[0], m[1], m[2]}))
			}
			return ids
		}
	}
	var ids []string
	for _, m := range censoredRe.FindAllStringSubmatch(s, 3) {
		if slices.Contains(notCodes, m[1]) {
			continue
		}
		num := m[2]
		if num == "" || len(num) < 3 { // zero padded ("ssis00001") or short ("abc-42")
			if num == "" {
				num = m[3]
			}
			n, _ := strconv.Atoi(num)
			num = fmt.Sprintf("%03d", n)
		}
		ids = append(ids, "jav-"+m[1]+"-"+num)
	}
	return ids
}

// aliasIndex finds a studio alias that is not part of a longer word.
func aliasIndex(s, alias string) int {
	for from := 0; ; {
		i := strings.Index(s[from:], alias)
		if i < 0 {
			return -1
		}
		i += from
		end := i + len(alias)
		if (i == 0 || !isAlnum(s[i-1])) && (end >= len(s) || !isLetter(s[end])) {
			return i
		}
		from = i + 1
	}
}

// adjacentMatch matches an ID directly after (or directly before) the studio
// alias, allowing only separators in between and no digits glued to it.
func adjacentMatch(re *regexp.Regexp, s string, after bool) []string {
	var loc []int
	if after {
		rest := strings.TrimLeft(s, "-_ .")
		if loc = re.FindStringSubmatchIndex(rest); loc == nil || loc[0] != 0 {
			return nil
		}
		s = rest
	} else {
		rest := strings.TrimRight(s, "-_ .[]()")
		all := re.FindAllStringSubmatchIndex(rest, -1)
		if len(all) == 0 || all[len(all)-1][1] != len(rest) {
			return nil
		}
		loc, s = all[len(all)-1], rest
		if loc[0] > 0 && isAlnum(s[loc[0]-1]) {
			return nil
		}
	}
	if loc[1] < len(s) && isDigit(s[loc[1]]) {
		return nil
	}
	m := make([]string, len(loc)/2)
	for i := range m {
		if loc[2*i] >= 0 {
			m[i] = s[loc[2*i]:loc[2*i+1]]
		}
	}
	return m
}

func isDigit(b byte) bool  { return b >= '0' && b <= '9' }
func isLetter(b byte) bool { return b >= 'a' && b <= 'z' }
func isAlnum(b byte) bool  { return isDigit(b) || isLetter(b) }

// sameCodeExact compares codes ignoring case and separators but not digits.
func sameCodeExact(a, b string) bool {
	return a != "" && nonAlnumRe.ReplaceAllString(strings.ToLower(a), "") == nonAlnumRe.ReplaceAllString(strings.ToLower(b), "")
}

// ---------- movies ----------

func (jm *JAVMania) GetMovieInfoByID(id string) (*model.MovieInfo, error) {
	u, err := movieURLFor(id)
	if err != nil {
		return nil, err
	}
	return jm.GetMovieInfoByURL(u)
}

func (jm *JAVMania) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := jm.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}
	homepage, err := movieURLFor(id) // English page, no language prefix or query
	if err != nil {
		return
	}
	slug, rest, _ := splitID(id)

	info = &model.MovieInfo{
		ID:            id,
		Provider:      jm.Name(),
		Homepage:      homepage,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}
	var contentID, finalURL string

	c := jm.ClonedCollector()

	c.OnResponse(func(r *colly.Response) { finalURL = r.Request.URL.String() })
	c.OnHTML(`h1`, func(e *colly.HTMLElement) {
		if info.Title == "" {
			info.Title = strings.TrimSpace(e.Text)
		}
	})
	c.OnHTML(`p.summary`, func(e *colly.HTMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})
	c.OnHTML(`div.tag-row .tag`, func(e *colly.HTMLElement) {
		if g := strings.TrimSpace(e.Text); g != "" && !slices.Contains(info.Genres, g) {
			info.Genres = append(info.Genres, g)
		}
	})
	c.OnHTML(`meta[property="og:image"]`, func(e *colly.HTMLElement) {
		if content := strings.TrimSpace(e.Attr("content")); content != "" {
			cover := e.Request.AbsoluteURL(content)
			info.CoverURL, info.BigCoverURL, info.ThumbURL = cover, cover, cover
		}
	})
	c.OnHTML(`div.gallery img`, func(e *colly.HTMLElement) {
		if src := strings.TrimSpace(e.Attr("src")); src != "" {
			if img := e.Request.AbsoluteURL(src); !slices.Contains(info.PreviewImages, img) {
				info.PreviewImages = append(info.PreviewImages, img)
			}
		}
	})
	c.OnHTML(`table.detail-table tr`, func(e *colly.HTMLElement) {
		th, td := e.DOM.Find("th").First(), e.DOM.Find("td").First()
		if th.Length() == 0 || td.Length() == 0 {
			return
		}
		node := td.Get(0)
		value := cellText(node)
		switch strings.TrimSpace(th.Text()) {
		case "Content ID":
			contentID = value
		case "Code":
			if fields := strings.Fields(value); len(fields) > 0 {
				contentID = fields[0]
			}
		case "Site", "Studio":
			info.Maker = value
		case "Label":
			info.Label = value
		case "Series":
			info.Series = value
		case "Director":
			info.Director = value
		case "Release date":
			info.ReleaseDate = parser.ParseDate(value)
		case "Runtime":
			info.Runtime = parser.ParseRuntime(value)
		case "Performer", "Performers":
			for _, name := range performerNames(node) {
				if !slices.Contains(info.Actors, name) {
					info.Actors = append(info.Actors, name)
				}
			}
		}
	})

	if err = c.Visit(homepage); err != nil {
		return nil, err
	}

	// Only accept the work that was asked for: challenge, landing and
	// redirected pages have another path or no matching code.
	if finalID, _ := jm.ParseMovieIDFromURL(finalURL); finalID != id {
		return nil, fmt.Errorf("javmania: %s redirected to %s", homepage, finalURL)
	}
	if contentID == "" || (slug == "jav" && !sameCodeExact(contentID, rest)) || (slug != "jav" && !strings.EqualFold(contentID, rest)) {
		return nil, fmt.Errorf("javmania: %s is not a work page for %s", homepage, id)
	}
	if slug == "jav" {
		info.Number = strings.ToUpper(contentID)
	} else {
		st, _ := siteBySlug(slug)
		info.Number = st.number(contentID)
	}
	// Titles end with "(Studio-ID)" or "(CODE)"; drop those suffixes.
	if loc := trailerRe.FindStringIndex(info.Title); loc != nil && strings.Contains(strings.ToLower(info.Title[loc[0]:]), strings.ToLower(contentID)) {
		if t := strings.TrimSpace(info.Title[:loc[0]]); t != "" {
			info.Title = t
		}
	}
	return info, nil
}

func (jm *JAVMania) NormalizeMovieKeyword(keyword string) string {
	if len(movieCandidates(keyword)) == 0 {
		return "" // not a JAV Mania title: skip this provider
	}
	return strings.ToLower(strings.TrimSpace(keyword))
}

func (jm *JAVMania) SearchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	for _, id := range movieCandidates(keyword) {
		info, e := jm.GetMovieInfoByID(id)
		if e != nil || !info.IsValid() {
			continue
		}
		results = append(results, info.ToSearchResult())
		if strings.HasPrefix(id, "jav-") {
			break // first matching censored code is enough
		}
	}
	return
}

// cellText is the visible text of a table cell without its Japanese spans.
func cellText(td *html.Node) string {
	var texts []string
	walkText(td, &texts)
	return strings.Join(strings.Fields(strings.Join(texts, " ")), " ")
}

func isJapaneseSpan(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "span" && htmlquery.SelectAttr(n, "lang") == "ja"
}

func walkText(n *html.Node, texts *[]string) {
	if isJapaneseSpan(n) {
		return
	}
	if n.Type == html.TextNode {
		*texts = append(*texts, n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkText(c, texts)
	}
}

// performerNames splits a performer cell: names are separated by <br> or ",".
func performerNames(td *html.Node) (names []string) {
	var buf []string
	flush := func() {
		for _, part := range strings.Split(strings.Join(buf, ""), ",") {
			if name := strings.Join(strings.Fields(part), " "); name != "" {
				names = append(names, name)
			}
		}
		buf = buf[:0]
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		switch {
		case n.Type == html.ElementNode && n.Data == "br":
			flush()
			return
		case isJapaneseSpan(n):
			return
		case n.Type == html.TextNode:
			buf = append(buf, n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(td)
	flush()
	return
}

// ---------- actors ----------
//
// Actor IDs are the JAV Mania actress slug, e.g. "yui-hatano".

func (jm *JAVMania) NormalizeActorID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if actorIDRe.MatchString(id) {
		return id
	}
	return ""
}

func (jm *JAVMania) ParseActorIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	m := actorPath.FindStringSubmatch(path.Clean(u.Path) + "/")
	if m == nil || !isAllowedHost(u) {
		return "", fmt.Errorf("javmania: not an actress page: %s", rawURL)
	}
	return m[1], nil
}

func (jm *JAVMania) GetActorInfoByID(id string) (*model.ActorInfo, error) {
	if id = jm.NormalizeActorID(id); id == "" {
		return nil, fmt.Errorf("javmania: invalid actor id")
	}
	return jm.GetActorInfoByURL(fmt.Sprintf(actorURL, id))
}

func (jm *JAVMania) GetActorInfoByURL(rawURL string) (info *model.ActorInfo, err error) {
	id, err := jm.ParseActorIDFromURL(rawURL)
	if err != nil {
		return
	}
	homepage := fmt.Sprintf(actorURL, id)
	info = &model.ActorInfo{
		ID:       id,
		Provider: jm.Name(),
		Homepage: homepage,
		Aliases:  []string{},
		Images:   []string{},
	}
	var finalURL, canonical string

	c := jm.ClonedCollector()
	c.OnResponse(func(r *colly.Response) { finalURL = r.Request.URL.String() })
	c.OnHTML(`link[rel="canonical"]`, func(e *colly.HTMLElement) { canonical = e.Attr("href") })
	c.OnHTML(`h1`, func(e *colly.HTMLElement) {
		if info.Name == "" {
			info.Name = strings.TrimSpace(e.Text)
		}
	})
	// "<name> (<span lang="ja">波多野結衣</span>) also appears on ..."
	c.OnHTML(`p`, func(e *colly.HTMLElement) {
		if !strings.Contains(e.Text, "also appears") {
			return
		}
		e.ForEach(`span[lang="ja"]`, func(_ int, s *colly.HTMLElement) {
			if ja := strings.TrimSpace(s.Text); ja != "" && !slices.Contains(info.Aliases, ja) {
				info.Aliases = append(info.Aliases, ja)
			}
		})
	})
	// "Profile: 163cm (5'4") / E-cup / B88 W59 H85 / Born 1988-05-24"
	c.OnHTML(`p.actress-profile`, func(e *colly.HTMLElement) {
		text := e.Text
		if m := profileRe.height.FindStringSubmatch(text); m != nil {
			info.Height = parser.ParseInt(m[1])
		}
		if m := profileRe.cup.FindStringSubmatch(text); m != nil {
			info.CupSize = m[1]
		}
		if m := profileRe.bwh.FindStringSubmatch(text); m != nil {
			info.Measurements = fmt.Sprintf("B:%s / W:%s / H:%s", m[1], m[2], m[3])
		}
		if m := profileRe.born.FindStringSubmatch(text); m != nil {
			info.Birthday = parser.ParseDate(m[1])
		}
	})

	if err = c.Visit(homepage); err != nil {
		return nil, err
	}
	if finalID, _ := jm.ParseActorIDFromURL(finalURL); finalID != id || canonical != homepage || info.Name == "" {
		return nil, fmt.Errorf("javmania: %s is not the actress page for %s", homepage, id)
	}
	return info, nil
}

// SearchActor accepts an English name ("Yui Hatano") or a Japanese name
// ("波多野結衣"); Japanese names are looked up on avhinban.com, which links
// each performer to her English JAV Mania page.
func (jm *JAVMania) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return
	}
	slug, japanese := slugify(keyword), false
	if slug == "" {
		slug, japanese = jm.slugFromJapaneseName(keyword), true
	}
	if slug == "" {
		return
	}
	info, e := jm.GetActorInfoByID(slug)
	if e != nil || !info.IsValid() || (japanese && !slices.Contains(info.Aliases, keyword)) {
		return // a link to another performer is never returned
	}
	return []*model.ActorSearchResult{info.ToSearchResult()}, nil
}

// slugify follows the site's slug rule for ASCII names; "" for other scripts.
func slugify(name string) string {
	for _, r := range name {
		if r > 127 {
			return ""
		}
	}
	return strings.Trim(nonAlnumRe.ReplaceAllString(strings.ToLower(name), "-"), "-")
}

// slugFromJapaneseName finds the one avhinban performer with exactly this
// name and follows her link to the English JAV Mania page.
func (jm *JAVMania) slugFromJapaneseName(name string) (slug string) {
	var performers []string
	search := jm.ClonedCollector()
	search.OnHTML(`a[href^="/performers/"]`, func(e *colly.HTMLElement) {
		if strings.TrimSpace(e.Text) == name && !slices.Contains(performers, e.Attr("href")) {
			performers = append(performers, e.Attr("href"))
		}
	})
	if err := search.Visit(fmt.Sprintf(avhinbanSearch, url.QueryEscape(name))); err != nil || len(performers) != 1 {
		return "" // unknown or ambiguous name: do not guess
	}

	page := avhinbanBaseURL + performers[0]
	var finalURL, link string
	profile := jm.ClonedCollector()
	profile.OnResponse(func(r *colly.Response) { finalURL = r.Request.URL.String() })
	profile.OnHTML(`dl.info`, func(e *colly.HTMLElement) {
		e.ForEach(`dt`, func(_ int, dt *colly.HTMLElement) {
			if strings.TrimSpace(dt.Text) == "英語ページ" {
				link = dt.DOM.Next().Find("a[href]").AttrOr("href", "")
			}
		})
	})
	if err := profile.Visit(page); err != nil || finalURL != page {
		return ""
	}
	slug, _ = jm.ParseActorIDFromURL(link)
	return slug
}

func init() {
	provider.Register(Name, New)
}
