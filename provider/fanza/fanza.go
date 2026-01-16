package fanza

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/docker/go-units"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"
	dt "gorm.io/datatypes"

	"github.com/metatube-community/metatube-sdk-go/collection/sets"
	"github.com/metatube-community/metatube-sdk-go/common/comparer"
	"github.com/metatube-community/metatube-sdk-go/common/js"
	"github.com/metatube-community/metatube-sdk-go/common/number"
	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/fanza/internal/graphql"
	"github.com/metatube-community/metatube-sdk-go/provider/fanza/internal/searchparse"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/imcmp"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var (
	_ provider.MovieProvider = (*FANZA)(nil)
	_ provider.MovieSearcher = (*FANZA)(nil)
	_ provider.MovieReviewer = (*FANZA)(nil)
)

const (
	Name     = "FANZA"
	Priority = 1000 + 1
)

const (
	baseURL                = "https://www.dmm.co.jp/"
	videoURL               = "https://video.dmm.co.jp/"
	baseDigitalURL         = "https://www.dmm.co.jp/digital/" // deprecated
	baseMonoURL            = "https://www.dmm.co.jp/mono/"
	searchURL              = "https://www.dmm.co.jp/search/=/searchstr=%s/limit=120/sort=date/"
	movieDigitalURL        = "https://video.dmm.co.jp/%s/content/?id=%s"
	movieDigitalAVURL      = "https://video.dmm.co.jp/av/content/?id=%s"
	movieDigitalAmateurURL = "https://video.dmm.co.jp/amateur/content/?id=%s"
	movieDigitalAnimeURL   = "https://video.dmm.co.jp/anime/content/?id=%s"
	movieDigitalCinemaURL  = "https://video.dmm.co.jp/cinema/content/?id=%s"
	movieMonoDVDURL        = "https://www.dmm.co.jp/mono/dvd/-/detail/=/cid=%s/"
	movieMonoAnimeURL      = "https://www.dmm.co.jp/mono/anime/-/detail/=/cid=%s/"
)

const regionNotAvailable = "not-available-in-your-region"

var (
	errContentIDNotFound = errors.New("content-id-not-found")
	errRequireNewHandler = errors.New("require-new-handler")

	ErrRegionNotAvailable = errors.New(regionNotAvailable)
)

type FANZA struct {
	*scraper.Scraper
	httpClient *http.Client
	videoAPI   *graphql.Client
}

func New() *FANZA {
	httpClient := &http.Client{}
	return &FANZA{
		httpClient: httpClient,
		videoAPI: graphql.NewClient(
			graphql.WithHTTPClient(httpClient),
		),
		Scraper: scraper.NewDefaultScraper(
			Name, baseURL, Priority, language.Japanese,
			scraper.WithCookies(baseURL, []*http.Cookie{
				{Name: "age_check_done", Value: "1"},
			}),
			scraper.WithCookies(videoURL, []*http.Cookie{
				{Name: "age_check_done", Value: "1"},
			}),
		),
	}
}

func (fz *FANZA) SetRequestTimeout(timeout time.Duration) {
	fz.httpClient.Timeout = timeout
	fz.Scraper.SetRequestTimeout(timeout)
}

func (fz *FANZA) NormalizeMovieID(id string) string {
	return strings.ToLower(id) /* FANZA uses lowercase ID */
}

func (fz *FANZA) getHomepagesByID(id string) []string {
	homepages := []string{
		fmt.Sprintf(movieMonoDVDURL, id),
		fmt.Sprintf(movieDigitalAVURL, id),
		fmt.Sprintf(movieDigitalAmateurURL, id),
		fmt.Sprintf(movieDigitalAnimeURL, id),
		fmt.Sprintf(movieMonoAnimeURL, id),
		fmt.Sprintf(movieDigitalCinemaURL, id),
	}
	if regexp.MustCompile(`(?i)[a-z]+00\d{3,}`).MatchString(id) {
		// might be digital av url, try it first.
		homepages[0], homepages[1] = homepages[1], homepages[0]
	}
	return homepages
}

func (fz *FANZA) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	for _, homepage := range fz.getHomepagesByID(id) {
		if info, err = fz.GetMovieInfoByURL(homepage); errors.Is(err, ErrRegionNotAvailable) || err == nil && info.IsValid() {
			return
		}
	}
	if err != nil && err.Error() == http.StatusText(http.StatusNotFound) {
		err = provider.ErrInfoNotFound
	}
	return
}

func (fz *FANZA) ParseMovieIDFromURL(rawURL string) (id string, err error) {
	defer func() {
		if err == nil && id == "" {
			err = errContentIDNotFound
		}
	}()
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if IsDigitalVideoURL(homepage.String()) {
		id = fz.NormalizeMovieID(
			homepage.Query().Get("id"))
	} else if sub := regexp.
		MustCompile(`/cid=(.*?)/`).
		FindStringSubmatch(homepage.Path); len(sub) == 2 {
		id = fz.NormalizeMovieID(sub[1])
	}
	return
}

func (fz *FANZA) GetMovieInfoByURL(rawURL string) (*model.MovieInfo, error) {
	if IsDigitalVideoURL(rawURL) {
		return fz.getDigitalMovieInfoByURL(rawURL)
	}
	return fz.getMonoMovieInfoByURL(rawURL)
}

func (fz *FANZA) getDigitalMovieInfoByURL(rawURL string) (*model.MovieInfo, error) {
	id, err := fz.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return nil, err
	}

	data, err := fz.videoAPI.GetContentPageData(id, graphql.BuildContentPageDataQueryOptions(rawURL))
	if err != nil {
		return nil, err
	}

	info := &model.MovieInfo{
		ID:            data.PPVContent.ID,
		Number:        data.PPVContent.MakerContentID,
		Title:         data.PPVContent.Title,
		Summary:       data.PPVContent.Description,
		Provider:      fz.Name(),
		Homepage:      fmt.Sprintf(movieDigitalURL, strings.ToLower(data.PPVContent.Floor), data.PPVContent.ID),
		ThumbURL:      data.PPVContent.PackageImage.MediumURL,
		CoverURL:      data.PPVContent.PackageImage.LargeURL,
		Maker:         data.PPVContent.Maker.Name,
		Label:         data.PPVContent.Label.Name,
		Series:        data.PPVContent.Series.Name,
		Runtime:       data.PPVContent.Duration / 60,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
		Score:         data.ReviewSummary.Average,
		ReleaseDate:   dt.Date(data.PPVContent.DeliveryStartDate),
	}

	// Director
	for _, director := range data.PPVContent.Directors {
		info.Director = director.Name
	}

	// Actors
	for _, actor := range data.PPVContent.Actresses {
		info.Actors = append(info.Actors, actor.Name)
	}

	// Actors (amateur fallback)
	if len(info.Actors) == 0 && data.PPVContent.AmateurActress.Name != "" {
		info.Actors = append(info.Actors, data.PPVContent.AmateurActress.Name)
	}

	// Genres
	for _, genre := range data.PPVContent.Genres {
		info.Genres = append(info.Genres, genre.Name)
	}

	// Preview Images
	for _, img := range data.PPVContent.SampleImages {
		info.PreviewImages = append(info.PreviewImages, img.LargeImageURL)
	}

	// Number (fallback)
	if info.Number == "" {
		info.Number = ParseNumber(info.ID)
	}

	// Release Date (fallback)
	if time.Time(info.ReleaseDate).IsZero() {
		info.ReleaseDate = dt.Date(data.PPVContent.MakerReleasedAt)
	}

	// Cover Image (fallback)
	if info.CoverURL == "" {
		info.CoverURL = data.PPVContent.PackageImage.MediumURL
	}

	// Big Thumb URL
	if info.BigThumbURL == "" {
		if fz.getImageSizeByURL(info.ThumbURL) > 100*units.KiB /* min big thumb size */ {
			info.BigThumbURL = info.ThumbURL
		}
	}

	// Big Thumb URL (fallback)
	if info.BigThumbURL == "" {
		fz.updateBigThumbURLFromPreviewImages(info)
	}

	// Preview Video
	if data.PPVContent.Sample2DMovie.HighestMovieURL != "" {
		info.PreviewVideoURL = data.PPVContent.Sample2DMovie.HighestMovieURL
	}

	// Preview Video (VR)
	if data.PPVContent.SampleVRMovie.HighestMovieURL != "" {
		info.PreviewVideoURL = data.PPVContent.SampleVRMovie.HighestMovieURL
	}

	return info, nil
}

func (fz *FANZA) getMonoMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := fz.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		Provider:      fz.Name(),
		Homepage:      rawURL,
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := fz.ClonedCollector()
	c.SetRedirectHandler(fz.digitalRedirectFunc)

	// Homepage
	c.OnRequest(func(r *colly.Request) {
		info.Homepage = r.URL.String()
	})

	// Title
	c.OnXML(`//*[@id="title"]`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Thumb
	c.OnXML(fmt.Sprintf(`//*[@id="package-src-%s"]`, id), func(e *colly.XMLElement) {
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Cover
	c.OnXML(fmt.Sprintf(`//*[@id="%s"]`, id), func(e *colly.XMLElement) {
		info.CoverURL = e.Request.AbsoluteURL(PreviewSrc(e.Attr("href")))
	})

	// Fields
	c.OnXML(`//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`.//td[1]`) {
		case "品番：":
			info.ID = e.ChildText(`.//td[2]`)
			info.Number = ParseNumber(info.ID)
		case "シリーズ：":
			info.Series = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "メーカー：":
			info.Maker = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "レーベル：":
			info.Label = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "ジャンル：":
			info.Genres = e.ChildTexts(`.//td[2]/a`)
		case "名前：":
			info.Actors = e.ChildTexts(`.//td[2]`)
		case "平均評価：":
			info.Score = fz.parseScoreFromURL(e.ChildAttr(`.//td[2]/img`, "src"))
		case "収録時間：":
			info.Runtime = parser.ParseRuntime(e.ChildText(`.//td[2]`))
		case "監督：":
			info.Director = strings.Trim(e.ChildText(`.//td[2]`), "-")
		case "配信開始日：", "商品発売日：", "発売日：", "貸出開始日：":
			if time.Time(info.ReleaseDate).IsZero() {
				info.ReleaseDate = parser.ParseDate(e.ChildText(`.//td[2]`))
			}
		}
	})

	// Actors (ajax)
	hasAPerformer := false
	c.OnXML(`//a[@id="a_performer"]`, func(e *colly.XMLElement) {
		hasAPerformer = true
	})

	// Actors (regular)
	c.OnXML(`//span[@id="performer"]`, func(e *colly.XMLElement) {
		parseActors(e.DOM.(*html.Node), (*[]string)(&info.Actors))
	})

	// JSON
	c.OnXML(`//script[@type="application/ld+json"]`, func(e *colly.XMLElement) {
		data := struct {
			Name        string `json:"name"`
			Image       string `json:"image"`
			Description string `json:"description"`
			Sku         string `json:"sku"`
			SubjectOf   struct {
				ContentUrl string `json:"contentUrl"`
				// EmbedUrl   string   `json:"embedUrl"`
				Genre []string `json:"genre"`
			} `json:"subjectOf"`
			AggregateRating struct {
				RatingValue string `json:"ratingValue"`
			} `json:"aggregateRating"`
		}{ /* assign default values */
			Name:        info.Title,
			Image:       info.ThumbURL,
			Description: info.Summary,
			Sku:         info.ID,
		}
		if json.Unmarshal([]byte(e.Text), &data) == nil {
			info.ID = data.Sku
			info.Number = ParseNumber(data.Sku)
			info.Title = data.Name
			info.Summary = data.Description
			info.ThumbURL = e.Request.AbsoluteURL(data.Image)
			if len(data.SubjectOf.Genre) > 0 {
				info.Genres = data.SubjectOf.Genre
			}
			if data.AggregateRating.RatingValue != "" {
				info.Score = parser.ParseScore(data.AggregateRating.RatingValue)
			}
			if data.SubjectOf.ContentUrl != "" {
				info.PreviewVideoURL = data.SubjectOf.ContentUrl
			}
		}
	})

	// Title (fallback)
	c.OnXML(`//meta[@property="og:title"]`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		info.Title = e.Attr("content")
	})

	// Summary (fallback)
	c.OnXML(`//div[@class="mg-b20 lh4"]`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		var summary string
		if summary = strings.TrimSpace(e.ChildText(`.//p[@class="mg-b20"]`)); summary != "" {
			// nop
		} else if summary = strings.TrimSpace(e.ChildText(`.//p`)); summary != "" {
			// nop
		} else {
			summary = strings.TrimSpace(e.Text)
		}
		info.Summary = summary
	})

	// Summary (incomplete fallback)
	c.OnXML(`//meta[@property="og:description"]`, func(e *colly.XMLElement) {
		if info.Summary != "" {
			return
		}
		info.Summary = e.Attr("content")
	})

	// Thumb (fallback)
	c.OnXML(`//*[@id="sample-video"]//img`, func(e *colly.XMLElement) {
		if info.ThumbURL != "" {
			return // ignore if not empty.
		}
		if !strings.HasPrefix(e.Attr("id"), "package-src") {
			return // probably not our img.
		}
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("src"))
	})

	// Cover (fallback)
	c.OnXML(`//*[@id="sample-video"]//a[@name="package-image"]`, func(e *colly.XMLElement) {
		if info.CoverURL != "" {
			return // ignore if not empty.
		}
		if href := e.Attr("href"); strings.HasSuffix(href, ".jpg") {
			info.CoverURL = e.Request.AbsoluteURL(href)
		}
	})

	// Thumb (fallback again)
	c.OnXML(`//meta[@property="og:image"]`, func(e *colly.XMLElement) {
		if info.ThumbURL != "" {
			return // ignore if not empty.
		}
		info.ThumbURL = e.Request.AbsoluteURL(e.Attr("content"))
	})

	// Preview Video (DVD)
	c.OnXML(`//*[@id="detail-sample-movie"]/div/a`, func(e *colly.XMLElement) {
		var videoPath string
		if dvu := e.Attr("data-video-url"); dvu != "" { // mono
			videoPath = dvu
		} else if v := e.Attr("onclick"); v != "" { // digital
			videoPath = regexp.MustCompile(`/(.+)/`).FindString(v)
		}
		info.PreviewVideoURL = fz.parsePreviewVideoURL(e.Request.AbsoluteURL(videoPath))
	})

	// Deprecated (?)
	// Preview Video (VR)
	c.OnXML(`//*[@id="detail-sample-vr-movie"]/div/a`, func(e *colly.XMLElement) {
		info.PreviewVideoURL = fz.parseVRPreviewVideoURL(
			e.Request.AbsoluteURL(
				regexp.MustCompile(`/(.+)/`).FindString(e.Attr("onclick"))))
	})

	// Preview Video (/digital/video*/)
	c.OnXML(`//script[@type="text/javascript"]`, func(e *colly.XMLElement) {
		// // プレイヤー呼び出し
		// if (autoPlayerMovieFlg) { // 通常プレイヤー
		// 	sampleplay(`/digital/${autoPlayerFloor}/-/detail/ajax-movie/=/cid=${gaContentId}/`);
		// } else { // VRプレイヤー
		// 	vrsampleplay(`/digital/-/vr-sample-player/=/cid=${gaContentId}/`);
		// }
		if jsCodes := regexp.MustCompile(`const\s+.+?\s*=.*?;`).FindAllString(e.Text, -1); len(jsCodes) > 0 {
			var (
				autoPlayerMovieFlg bool
				autoPlayerFloor    string
			)
			extractCode := func(v string) string {
				for _, jsCode := range jsCodes {
					if strings.Contains(jsCode, v) {
						return strings.ReplaceAll(jsCode, "const ", "var ")
					}
				}
				return ""
			}
			_ = js.UnmarshalObject(extractCode("autoPlayerFloor"), "autoPlayerFloor", &autoPlayerFloor)
			_ = js.UnmarshalObject(extractCode("autoPlayerMovieFlg"), "autoPlayerMovieFlg", &autoPlayerMovieFlg)
			if autoPlayerFloor == "" {
				return // skip
			}
			if autoPlayerMovieFlg {
				sampleURL := e.Request.AbsoluteURL(fmt.Sprintf(`/digital/%s/-/detail/ajax-movie/=/cid=%s/`, autoPlayerFloor, info.ID))
				info.PreviewVideoURL = fz.parsePreviewVideoURL(sampleURL)
			} else {
				vrSampleURL := e.Request.AbsoluteURL(fmt.Sprintf(`/digital/-/vr-sample-player/=/cid=%s/`, info.ID))
				info.PreviewVideoURL = fz.parseVRPreviewVideoURL(vrSampleURL)
			}
		}
	})

	// In case of any duplication
	previewImageSet := sets.NewOrderedSet[string]()
	extractImageSrc := func(e *colly.XMLElement) string {
		src := e.ChildAttr(`.//img`, "data-lazy")
		if strings.TrimSpace(src) == "" {
			src = e.ChildAttr(`.//img`, "src")
		}
		return src
	}

	// Preview Images Digital/DVD
	c.OnXML(`//*[@id="sample-image-block"]//a[@name="sample-image"]`, func(e *colly.XMLElement) {
		previewImageSet.Add(e.Request.AbsoluteURL(PreviewSrc(extractImageSrc(e))))
	})

	// Preview Images Digital (Fallback)
	c.OnXML(`//*[@id="sample-image-block"]/a`, func(e *colly.XMLElement) {
		if previewImageSet.Len() == 0 {
			return
		}
		previewImageSet.Add(e.Request.AbsoluteURL(PreviewSrc(extractImageSrc(e))))
	})

	// Final Preview Images
	c.OnScraped(func(_ *colly.Response) {
		info.PreviewImages = previewImageSet.AsSlice()
	})

	// Final (images)
	c.OnScraped(func(_ *colly.Response) {
		if info.CoverURL == "" {
			// try to convert thumb url to cover url.
			info.CoverURL = PreviewSrc(info.ThumbURL)
		}
	})

	// Final (big thumb image)
	c.OnScraped(func(_ *colly.Response) {
		fz.updateBigThumbURLFromPreviewImages(info)
	})

	// Final (actors)
	c.OnScraped(func(r *colly.Response) {
		if !hasAPerformer {
			return
		}

		n, innerErr := htmlquery.Parse(bytes.NewReader(r.Body))
		if innerErr != nil {
			return
		}

		n = htmlquery.FindOne(n, `//script[contains(text(),"a#a_performer")]/text()`)
		if n == nil {
			return
		}

		if ss := regexp.MustCompile(`url:\s*'(.+?)',`).
			FindStringSubmatch(n.Data); len(ss) == 2 && strings.TrimSpace(ss[1]) != "" {
			d := c.Clone()
			d.OnXML(`/` /* root */, func(e *colly.XMLElement) {
				var actors []string
				parseActors(e.DOM.(*html.Node), &actors)
				if len(actors) > 0 {
					info.Actors = actors // replace with new actors.
				}
			})
			d.Visit(r.Request.AbsoluteURL(ss[1]))
		}
	})

	c.OnScraped(func(r *colly.Response) {
		if !info.IsValid() && isRegionError(r) {
			err = ErrRegionNotAvailable
		}
	})

	vErr := c.Visit(info.Homepage)
	if vErr != nil {
		var urlErr *url.Error
		if errors.As(vErr, &urlErr) && errors.Is(urlErr.Err, errRequireNewHandler) {
			return fz.getDigitalMovieInfoByURL(urlErr.URL) // use the new handler.
		}
		err = vErr
	}
	return
}

func (fz *FANZA) NormalizeMovieKeyword(keyword string) string {
	if number.IsSpecial(keyword) {
		return ""
	}
	return strings.ToLower(keyword) /* FANZA prefers lowercase */
}

func (fz *FANZA) SearchMovie(keyword string) ([]*model.MovieSearchResult, error) {
	if strings.Contains(keyword, "-") {
		if results, err := fz.searchMovieNext(strings.Replace(keyword,
			/* FANZA cannot search hyphened number */
			"-", "00", 1) +
			/* Add a `#` sign to distinguish 001 style number */
			"#"); err == nil && len(results) > 0 {
			return results, nil
		}
	}
	// fallback to normal dvd search.
	return fz.searchMovieNext(strings.Replace(keyword, "-", "", 1))
}

func (fz *FANZA) searchMovieNext(keyword string) (results []*model.MovieSearchResult, err error) {
	defer func() {
		fz.sortMovieSearchResults(keyword, results)
	}()

	c := fz.ClonedCollector()
	p := searchparse.NewSearchPageParser()

	c.OnXML("//script", func(e *colly.XMLElement) {
		_ = p.LoadJSCode(e.Text)
	})

	if err = c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword))); err != nil {
		return
	}

	resp := &searchparse.ResponseWrapper{}
	if err = p.Parse(resp); err != nil {
		return
	}

	filter := func(url string) bool {
		for _, prefix := range []string{
			baseDigitalURL,
			baseMonoURL,
			videoURL,
		} {
			if strings.HasPrefix(url, prefix) {
				return true
			}
		}
		return false
	}

	for _, product := range resp.BackendResponse.Contents.Data {
		if !filter(product.DetailURL) {
			continue // ignore non-digital/mono results, e.g.: 月額動画
		}
		var releaseDate string
		if re := regexp.MustCompile(`(配信日|発売日|貸出日)：\s*`); re.MatchString(product.ReleaseAnnouncement) {
			releaseDate = re.ReplaceAllString(product.ReleaseAnnouncement, "")
		}
		results = append(results, &model.MovieSearchResult{
			ID:          product.ContentID,
			Number:      ParseNumber(product.ContentID),
			Title:       product.Title,
			Provider:    fz.Name(),
			Homepage:    product.DetailURL,
			Actors:      product.Casts,
			ThumbURL:    product.ThumbnailImageURL,
			CoverURL:    PreviewSrc(product.ThumbnailImageURL),
			Score:       product.Rate,
			ReleaseDate: parser.ParseDate(releaseDate /* 発売日：2022/07/21 */),
		})
	}
	return
}

// Deprecated: this function is deprecated.
//
//nolint:unused // ignore unused warning for this function.
func (fz *FANZA) searchMovie(keyword string) (results []*model.MovieSearchResult, err error) {
	defer func() {
		fz.sortMovieSearchResults(keyword, results)
	}()

	c := fz.ClonedCollector()

	c.OnXML(`//*[@id="list"]/li`, func(e *colly.XMLElement) {
		homepage := e.Request.AbsoluteURL(e.ChildAttr(`.//p[@class="tmb"]/a`, "href"))
		if !strings.HasPrefix(homepage, baseDigitalURL) &&
			!strings.HasPrefix(homepage, baseMonoURL) &&
			!strings.HasPrefix(homepage, videoURL) {
			return // ignore other contents.
		}
		id, _ := fz.ParseMovieIDFromURL(homepage) // ignore error.

		thumb := e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "src")
		if re := regexp.MustCompile(`(p[a-z]\.)jpg`); re.MatchString(thumb) {
			thumb = re.ReplaceAllString(thumb, "ps.jpg")
		}

		var releaseDate string
		rate := e.ChildText(`.//p[@class="rate"]`)
		if re := regexp.MustCompile(`(配信日|発売日|貸出日)：\s*`); re.MatchString(rate) {
			releaseDate = re.ReplaceAllString(rate, "")
			rate = "" // reset rate.
		}
		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Number:      ParseNumber(id),
			Title:       e.ChildAttr(`.//p[@class="tmb"]/a/span[1]/img`, "alt"),
			Provider:    fz.Name(),
			Homepage:    homepage,
			ThumbURL:    e.Request.AbsoluteURL(thumb),
			CoverURL:    e.Request.AbsoluteURL(PreviewSrc(thumb)),
			Score:       parser.ParseScore(rate /* float or a dash (-) */),
			ReleaseDate: parser.ParseDate(releaseDate /* 発売日：2022/07/21 */),
		})
	})

	c.OnScraped(func(r *colly.Response) {
		if isRegionError(r) {
			err = ErrRegionNotAvailable
		}
	})

	if vErr := c.Visit(fmt.Sprintf(searchURL, url.QueryEscape(keyword))); vErr != nil {
		err = vErr
	}
	return
}

func (fz *FANZA) sortMovieSearchResults(keyword string, results []*model.MovieSearchResult) {
	if len(results) == 0 {
		return
	}
	r := regexp.MustCompile(`(?i)([A-Z]+)0*([1-9]*)`)
	x := r.ReplaceAllString(keyword, "${1}${2}")
	sort.SliceStable(results, func(i, j int) bool {
		a := r.ReplaceAllString(results[i].ID, "${1}${2}")
		b := r.ReplaceAllString(results[j].ID, "${1}${2}")
		if a == b {
			// prefer digital results.
			return strings.Contains(results[i].Homepage, "/digital/") ||
				!strings.Contains(results[j].Homepage, "/digital/")
		}
		return comparer.Compare(a, x) >= comparer.Compare(b, x)
	})
}

func (fz *FANZA) GetMovieReviewsByID(id string) (reviews []*model.MovieReviewDetail, err error) {
	for _, homepage := range fz.getHomepagesByID(id) {
		if reviews, err = fz.GetMovieReviewsByURL(homepage); err == nil && len(reviews) > 0 {
			return
		}
	}
	return nil, provider.ErrInfoNotFound
}

func (fz *FANZA) GetMovieReviewsByURL(rawURL string) (reviews []*model.MovieReviewDetail, err error) {
	if IsDigitalVideoURL(rawURL) {
		return fz.getDigitalMovieReviewsByURL(rawURL)
	}
	return fz.getMonoMovieReviewsByURL(rawURL)
}

func (fz *FANZA) getDigitalMovieReviewsByURL(rawURL string) (reviews []*model.MovieReviewDetail, err error) {
	id, err := fz.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	data, err := fz.videoAPI.GetUserReviews(id)
	if err != nil {
		return
	}

	for _, review := range data.Reviews.Items {
		reviews = append(reviews, &model.MovieReviewDetail{
			Title:   review.Title,
			Author:  review.Nickname,
			Comment: review.Comment,
			Score:   float64(review.Rating),
			Date:    dt.Date(review.PublishDate),
		})
	}
	return
}

func (fz *FANZA) getMonoMovieReviewsByURL(rawURL string) (reviews []*model.MovieReviewDetail, err error) {
	c := fz.ClonedCollector()
	c.SetRedirectHandler(fz.digitalRedirectFunc)

	c.OnXML(`//*[starts-with(@id, 'review')]//div[ends-with(@class, 'review__list')]/ul/li`, func(e *colly.XMLElement) {
		comment := strings.TrimSpace(e.ChildText(`.//div[1]`))

		var name string
		if n := htmlquery.FindOne(e.DOM.(*html.Node), `.//div[2]/p/span[ends-with(@class, 'review__unit__reviewer')]/a`); n != nil {
			if n := n.FirstChild; n != nil && n.Type == html.TextNode {
				name = strings.TrimSpace(n.Data)
			}
		}
		if name == "" /* fallback */ {
			name = strings.TrimSpace(regexp.MustCompile(`(さん)?(のレビュー)?`).ReplaceAllString(
				e.ChildText(`.//div[2]/p/span[ends-with(@class, 'review__unit__reviewer')]`), ""))
		}

		if name == "" || comment == "" {
			return
		}

		score := 0.0
		ratings := strings.Split(strings.TrimSpace(e.ChildAttr(`.//p/span[1]`, "class")), "-")
		if len(ratings) > 0 {
			score = parser.ParseScore(ratings[len(ratings)-1]) / 10
			if score > 5.0 {
				score = 0 // reset, must be an error
			}
		}

		reviews = append(reviews, &model.MovieReviewDetail{
			Author:  name,
			Comment: comment,
			Score:   score,
			Title:   strings.TrimSpace(e.ChildText(`.//p/span[ends-with(@class, 'review__unit__title')]`)),
			Date: parser.ParseDate(strings.Trim(
				e.ChildText(`.//div[2]/p/span[ends-with(@class, 'review__unit__postdate')]`), "- ")),
		})
	})

	if err = c.Visit(rawURL); err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) && errors.Is(urlErr.Err, errRequireNewHandler) {
			return fz.getDigitalMovieReviewsByURL(urlErr.URL)
		}
	}
	return
}

func parseActors(n *html.Node, texts *[]string) {
	if n.Type == html.TextNode {
		// custom trim function.
		if text := strings.Trim(strings.TrimSpace(n.Data), "-/"); text != "" {
			*texts = append(*texts, text)
		}
	}
	for n := n.FirstChild; n != nil; n = n.NextSibling {
		// handle `id="a_performer"` situation.
		if n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "a_performer" {
					goto next
				}
			}
		}
		parseActors(n, texts)
	next:
		continue
	}
}

func isRegionError(r *colly.Response) bool {
	const accountsDomain = "accounts.dmm.co.jp"
	if strings.Contains(r.Request.URL.Path, regionNotAvailable) ||
		strings.Contains(r.Request.URL.Host, accountsDomain) {
		return true
	}
	return false
}

func (fz *FANZA) parseScoreFromURL(s string) float64 {
	u, err := url.Parse(s)
	if err != nil {
		return 0
	}
	gif := path.Base(u.Path)
	ext := path.Ext(gif)
	n := gif[:len(gif)-len(ext)]
	score, _ := strconv.ParseFloat(n, 64)
	if score > 5.0 {
		// Fix scores for mono/anime.
		// e.g.: https://review.dmm.com/web/images/pc/45.gif
		score = score / 10.0
	}
	return score
}

// Deprecated: this is unneeded.
//
//nolint:unused // ignore unused warning for this function.
func (fz *FANZA) updateWithAWSImgSrc(info *model.MovieInfo) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if time.Time(info.ReleaseDate).Before(start) {
		return // ignore movies released before this date.
	}
	if !strings.Contains(info.Homepage, "/digital/videoa") {
		return // ignore non-digital/videoa typed movies.
	}
	c := fz.ClonedCollector()
	c.Async = true
	c.ParseHTTPErrorResponse = false
	c.OnResponseHeaders(func(r *colly.Response) {
		if r.Headers.Get("Content-Type") != "image/jpeg" {
			return // ignore non-image/jpeg contents.
		}
		length, _ := strconv.Atoi(r.Headers.Get("Content-Length"))
		switch {
		case strings.HasSuffix(info.ThumbURL, path.Base(r.Request.URL.Path)) && length > 100*units.KiB:
			info.BigThumbURL = r.Request.URL.String()
		case strings.HasSuffix(info.CoverURL, path.Base(r.Request.URL.Path)) && length > 500*units.KiB:
			info.BigCoverURL = r.Request.URL.String()
		}
		// abort to prevent image content from being downloaded.
		r.Request.Abort()
	})
	c.Visit(strings.ReplaceAll(info.ThumbURL,
		"https://pics.dmm.co.jp/",
		"https://awsimgsrc.dmm.co.jp/pics_dig/"))
	c.Visit(strings.ReplaceAll(info.CoverURL,
		"https://pics.dmm.co.jp/",
		"https://awsimgsrc.dmm.co.jp/pics_dig/"))
	c.Wait()
}

// getImageSizeByURL retrieves the image size from the Content-Length header of a given URL.
func (fz *FANZA) getImageSizeByURL(imgURL string) (size int) {
	c := fz.ClonedCollector()
	c.OnResponseHeaders(func(r *colly.Response) {
		if !strings.HasPrefix(r.Headers.Get("Content-Type"), "image/") {
			return // ignore non-image content.
		}
		size, _ = strconv.Atoi(r.Headers.Get("Content-Length"))
	})
	c.Visit(imgURL)
	return
}

// updateBigThumbURLFromPreviewImages attempts to update the big thumb
// image URL with the first preview image if the two images match.
func (fz *FANZA) updateBigThumbURLFromPreviewImages(info *model.MovieInfo) {
	if info.BigThumbURL != "" /* a big thumb already exists */ ||
		info.ThumbURL == "" /* thumb url is empty */ ||
		len(info.PreviewImages) == 0 /* no preview images */ {
		return
	}

	if imcmp.Similar(info.ThumbURL, info.PreviewImages[0], nil) {
		// populate the first preview image as a big thumb image.
		info.BigThumbURL = info.PreviewImages[0]
		info.PreviewImages = info.PreviewImages[1:]
		return
	}
}

func (fz *FANZA) parsePreviewVideoURL(videoURL string) (previewVideoURL string) {
	c := fz.ClonedCollector()
	// In case it's an iframe page:
	// E.g.: https://www.dmm.co.jp/digital/videoa/-/detail/ajax-movie/=/cid=1start00190/
	c.OnXML(`//iframe`, func(e *colly.XMLElement) {
		previewVideoURL = fz.parsePreviewVideoURL(
			e.Request.AbsoluteURL(e.Attr("src")),
		)
	})
	// E.g.: https://www.dmm.co.jp/service/digitalapi/-/html5_player/=/cid=1start00190/
	c.OnResponse(func(r *colly.Response) {
		if resp := regexp.MustCompile(`const args = (\{.+});`).FindSubmatch(r.Body); len(resp) == 2 {
			data := struct {
				Bitrates []struct {
					// Bitrate int    `json:"bitrate"`
					Src string `json:"src"`
				} `json:"bitrates"`
			}{}
			if json.Unmarshal(resp[1], &data) == nil && len(data.Bitrates) > 0 {
				previewVideoURL = r.Request.AbsoluteURL(data.Bitrates[0].Src)
			}
		}
	})
	c.Visit(videoURL)
	return
}

func (fz *FANZA) parseVRPreviewVideoURL(vrVideoURL string) (previewVideoURL string) {
	c := fz.ClonedCollector()
	c.OnResponse(func(r *colly.Response) {
		sub := regexp.MustCompile(`var sampleUrl = "(.+?)";`).FindSubmatch(r.Body)
		if len(sub) == 2 {
			previewVideoURL = r.Request.AbsoluteURL(string(sub[1]))
		}
	})
	c.Visit(vrVideoURL)
	return
}

func (fz *FANZA) digitalRedirectFunc(req *http.Request, _ []*http.Request) error {
	if strings.HasPrefix(req.URL.String(), videoURL) {
		if !strings.HasSuffix(req.URL.Path, "/") {
			// ensure trailing slash.
			req.URL.Path += "/"
		}
		return &url.Error{
			Op:  "redirect to",
			URL: req.URL.String(),
			Err: errRequireNewHandler,
		}
	}
	return nil
}

func IsDigitalVideoURL(url string) bool {
	return strings.HasPrefix(url, videoURL)
}

// ParseNumber parses FANZA-formatted id to general ID.
func ParseNumber(s string) string {
	if ss := regexp.MustCompile(`([A-Z]+)(\d+)`).FindStringSubmatch(strings.ToUpper(s)); len(ss) >= 3 {
		n, _ := strconv.Atoi(ss[2])
		return fmt.Sprintf("%s-%03d", ss[1], n)
	}
	return s
}

// PreviewSrc maximize the preview image.
// Ref: https://digstatic.dmm.com/js/digital/preview_jquery.js#652
// JS Code:
// // 画像パスの正規化
// function preview_src(src)
//
//	{
//		  if (src.match(/(p[a-z]\.)jpg/)) {
//			  return src.replace(RegExp.$1, 'pl.');
//		  } else if (src.match(/consumer_game/)) {
//			  return src.replace('js-','-');
//		  } else if (src.match(/js\-([0-9]+)\.jpg$/)) {
//			  return src.replace('js-','jp-');
//		  } else if (src.match(/ts\-([0-9]+)\.jpg$/)) {
//			  return src.replace('ts-','tl-');
//		  } else if (src.match(/(\-[0-9]+\.)jpg$/)) {
//			  return src.replace(RegExp.$1, 'jp' + RegExp.$1);
//		  } else {
//			  return src.replace('-','jp-');
//		  }
//	}
func PreviewSrc(s string) string {
	if re := regexp.MustCompile(`(p[a-z]\.)jpg`); re.MatchString(s) {
		return re.ReplaceAllString(s, "pl.jpg")
	} else if re = regexp.MustCompile(`consumer_game`); re.MatchString(s) {
		return strings.ReplaceAll(s, "js-", "-")
	} else if re = regexp.MustCompile(`js-(\d+)\.jpg$`); re.MatchString(s) {
		return strings.ReplaceAll(s, "js-", "jp-")
	} else if re = regexp.MustCompile(`ts-(\d+)\.jpg$`); re.MatchString(s) {
		return strings.ReplaceAll(s, "ts-", "tl-")
	} else if re = regexp.MustCompile(`(-\d+\.)jpg$`); re.MatchString(s) {
		return re.ReplaceAllString(s, "jp${1}jpg")
	} else {
		return strings.ReplaceAll(s, "-", "jp-")
	}
}

func init() {
	provider.Register(Name, New)
}
