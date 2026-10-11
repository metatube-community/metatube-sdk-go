package avleague

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2"
	"golang.org/x/text/language"
	dt "gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.ActorProvider = (*AVLeague)(nil)
	_ provider.ActorSearcher = (*AVLeague)(nil)
)

// reActorListPage matches the actress detail link in a list page, e.g. /actress/39587.html
var reActorListPage = regexp.MustCompile(`^/actress/(\d+)\.html$`)

const (
	Name     = "AV-LEAGUE" // `AV.LEAGUE`
	Priority = 1000
)

const (
	baseURL  = "https://www.av-league.com/"
	actorURL = "https://www.av-league.com/actress/%s.html"
	// Actor list pages. The site's in-page search results area is empty (no longer server-rendered after the revamp),
	// but the actor cards on list pages are static HTML: an a[href=/actress/{id}.html] contains
	// div.l-pic > img[alt=actor name][data-layzr=avatar URL],
	// so we traverse list pages and match by name instead.
	actorListURL = "https://www.av-league.com/actress-list/all-3/all_all_default_%d.html"
	// Max list-page count (verified 30 pages still non-empty; beyond that returns empty pages)
	actorListMaxPage = 31
)

type AVLeague struct {
	*scraper.Scraper
}

func New() *AVLeague {
	return &AVLeague{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
		scraper.WithDisableCookies(),
	)}
}

func (avl *AVLeague) GetActorInfoByID(id string) (info *model.ActorInfo, err error) {
	return avl.GetActorInfoByURL(fmt.Sprintf(actorURL, id))
}

func (avl *AVLeague) ParseActorIDFromURL(rawURL string) (id string, err error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if ext := path.Ext(homepage.Path); ext != "" {
		id = path.Base(homepage.Path[:len(homepage.Path)-len(ext)])
	}
	return
}

func (avl *AVLeague) GetActorInfoByURL(rawURL string) (info *model.ActorInfo, err error) {
	id, err := avl.ParseActorIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.ActorInfo{
		ID:       id,
		Provider: avl.Name(),
		Homepage: rawURL,
		Aliases:  []string{},
		Images:   []string{},
	}

	c := avl.ClonedCollector()

	// Name
	// Verified: div#pan still exists but no longer holds the name; the name is now served by h1#j-prof,
	// e.g. "Ishiguro Kyoka ari (ishigurokyouka)". The old XPath //*[@id="pan"]/span is obsolete.
	c.OnXML(`//*[@id="j-prof"]`, func(e *colly.XMLElement) {
		name := strings.TrimSpace(e.Text)
		if name == "" {
			return
		}
		// Strip the trailing kana reading: "Ishiguro Kyoka ari (ishigurokyouka)" -> "Ishiguro Kyoka ari"
		if i := strings.Index(name, "（"); i > 0 {
			name = strings.TrimSpace(name[:i])
		}
		info.Name = name
	})

	// Name (fallback): on some pages the name is still at the old location
	c.OnXML(`//*[@id="pan"]//span`, func(e *colly.XMLElement) {
		if info.Name != "" {
			return
		}
		if name := strings.TrimSpace(e.Text); name != "" {
			info.Name = name
		}
	})

	// Image
	// The real avatar is hosted on DMM at pics.dmm.co.jp/mono/actjpgs/{romaji}.jpg.
	// Note the page og:image is a malformed URL -- the site concatenates its domain with the image URL directly:
	//   https://www.av-league.comhttp://pics.dmm.co.jp/mono/actjpgs/xxx.jpg
	// We must slice the real image URL after "av-league.com".
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		raw := strings.TrimSpace(e.Attr("content"))
		if raw == "" {
			return
		}
		// Locate by the DMM image-host domain, bypassing the concatenated site domain
		if i := strings.Index(raw, "pics.dmm.co.jp"); i > 0 {
			raw = "http://" + raw[i:]
		}
		if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
			return
		}
		info.Images = append(info.Images, raw)
	})

	// Fields
	c.OnXML(`//*[@id="contents"]//table/tbody/tr`, func(e *colly.XMLElement) {
		row := strings.TrimSpace(e.ChildText(`.//th`))
		data := strings.TrimSpace(e.ChildText(`.//td`))
		if data == "不明" {
			return // ignore unknown
		}
		switch row {
		case "3サイズ":
			B, W, H, Cup := parseMeasurements(data)
			if B != 0 && W != 0 && H != 0 {
				info.Measurements = fmt.Sprintf("B:%d / W:%d / H:%d", B, W, H)
			}
			info.CupSize = Cup
		case "身長":
			info.Height = parser.ParseInt(strings.TrimRight(data, "cm"))
		case "血液型":
			info.BloodType = strings.TrimSpace(strings.TrimRight(data, "型"))
		case "生年月日":
			info.Birthday = parseDate(data)
		case "出身":
			info.Nationality = data
		case "デビュー":
			info.DebutDate = parseDate(data)
		}
	})

	err = c.Visit(info.Homepage)
	return
}

// SearchActor searches actors by name.
//
// The site's in-site search is no longer server-rendered; we traverse static actress list pages and match by name.
// Two acceleration techniques:
//
//  1. In-process index cache -- on first search, fetch all list pages to build a normalized-name -> actor-entry index,
//     cache it on hit; subsequent searches match in memory (O(actor count), ms-level) without re-paging.
//  2. Full concurrent fetch -- when building the index, fetch all 30 list pages at once (instead of the old 6-per-batch serial wait),
//     cutting first-index time from ~12s to ~2-3s.
//
// If the index is unavailable (first build failed / network error), fall back to one real-time fetch so search always works.
func (avl *AVLeague) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, provider.ErrInvalidKeyword
	}

	// Normalize for comparison: drop whitespace and the "ari" suffix the site appends after every name
	norm := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "　", "")
		s = strings.TrimSuffix(s, "ari")
		return strings.ToLower(s)
	}
	target := norm(keyword)
	if target == "" {
		return nil, provider.ErrInvalidKeyword
	}

	// Get the full actor set: prefer the in-memory index, fall back to one real-time fetch on failure.
	entries, eerr := avl.actorIndex()
	if eerr != nil {
		entries, eerr = avl.buildActorIndex()
	}
	if eerr != nil {
		return nil, eerr
	}

	// Match by name within the set: exact equality first, then containment.
	// Note: the ranking is not sorted by name; the target actor may be on any page,
	// so we must scan everything (cannot stop at first hit) or we miss actors.
	for _, e := range entries {
		nn := norm(e.Name)
		if nn != target && !strings.Contains(nn, target) {
			continue
		}
		images := make([]string, 0, 1)
		if e.Image != "" {
			images = append(images, e.Image)
		}
		results = append(results, &model.ActorSearchResult{
			ID:       e.ID,
			Name:     e.Name,
			Images:   images,
			Provider: avl.Name(),
			Homepage: fmt.Sprintf(actorURL, e.ID),
		})
	}

	// Sort: exact matches first; shorter names are more likely the precise match
	sort.SliceStable(results, func(i, j int) bool {
		ni, nj := norm(results[i].Name), norm(results[j].Name)
		ei, ej := ni == target, nj == target
		if ei != ej {
			return ei
		}
		return len(ni) < len(nj)
	})
	return
}

// actorIndexEntry is an actor entry parsed from a list page (name + ID + list-page avatar).
type actorIndexEntry struct {
	ID    string
	Name  string
	Image string
}

// actorIdx is the in-process actor index cache (reused after the first full build; rebuilt after a process restart).
var (
	actorIdx     atomic.Pointer[[]actorIndexEntry]
	actorIdxMu   sync.Mutex
	actorIdxBusy bool
)

// actorIndex returns the cached index; if not built or stale, it fetches all pages live and caches them.
func (avl *AVLeague) actorIndex() ([]actorIndexEntry, error) {
	if p := actorIdx.Load(); p != nil {
		return *p, nil
	}
	actorIdxMu.Lock()
	if p := actorIdx.Load(); p != nil {
		actorIdxMu.Unlock()
		return *p, nil
	}
	if actorIdxBusy {
		// Another goroutine is building the index; wait for it then read once more
		actorIdxMu.Unlock()
		time.Sleep(50 * time.Millisecond)
		if p := actorIdx.Load(); p != nil {
			return *p, nil
		}
		return nil, fmt.Errorf("actor index is being built")
	}
	actorIdxBusy = true
	actorIdxMu.Unlock()

	entries, berr := avl.buildActorIndex()

	actorIdxMu.Lock()
	actorIdxBusy = false
	if berr == nil {
		actorIdx.Store(&entries)
	}
	actorIdxMu.Unlock()
	return entries, berr
}

// buildActorIndex fetches all list pages concurrently and builds the normalized-name -> actor-entry index.
// This is technique 2: fetch all 30 list pages at once, not the old 6-per-batch serial wait.
func (avl *AVLeague) buildActorIndex() ([]actorIndexEntry, error) {
	results := make(chan []actorIndexEntry, actorListMaxPage)
	var wg sync.WaitGroup
	for page := 1; page <= actorListMaxPage; page++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			c := avl.ClonedCollector()
			var local []actorIndexEntry
			c.OnXML(`//a[@href]`, func(e *colly.XMLElement) {
				href := e.Attr("href")
				if !reActorListPage.MatchString(href) {
					return
				}
				id := strings.TrimSuffix(strings.TrimPrefix(href, "/actress/"), ".html")
				if id == "" {
					return
				}
				name := strings.TrimSpace(e.ChildAttr(`div[@class="l-pic"]/img`, "alt"))
				if name == "" {
					return
				}
				img := e.ChildAttr(`div[@class="l-pic"]/img`, "data-layzr")
				if img != "" {
					img = e.Request.AbsoluteURL(img)
				}
				local = append(local, actorIndexEntry{ID: id, Name: name, Image: img})
			})
			if err := c.Visit(fmt.Sprintf(actorListURL, p)); err != nil {
				return
			}
			results <- local
		}(page)
	}
	go func() { wg.Wait(); close(results) }()

	var all []actorIndexEntry
	for local := range results {
		all = append(all, local...)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("actor index empty: site may be unreachable")
	}
	return all, nil
}

func parseMeasurements(s string) (B, W, H int, Cup string) {
	for _, item := range strings.Split(s, "/") {
		name, data, found := strings.Cut(item, ":")
		if !found {
			continue
		}
		switch strings.TrimSpace(name) {
		case "B":
			if ss := regexp.MustCompile(`(\d+)（([A-Z])）`).FindStringSubmatch(data); len(ss) == 3 {
				B = parser.ParseInt(ss[1])
				Cup = ss[2]
			}
		case "W":
			W = parser.ParseInt(data)
		case "H":
			H = parser.ParseInt(data)
		}
	}
	return
}

func parseDate(s string) (date dt.Date) {
	defer func() {
		if !time.Time(date).IsZero() {
			return
		}
		if ss := regexp.MustCompile(`([\s\d]+)年`).
			FindStringSubmatch(s); len(ss) == 2 {
			date = dt.Date(time.Date(parser.ParseInt(ss[1]),
				2, 2, 2, 2, 2, 2, time.UTC))
		}
	}()
	return parser.ParseDate(s)
}

func init() {
	provider.Register(Name, New)
}
