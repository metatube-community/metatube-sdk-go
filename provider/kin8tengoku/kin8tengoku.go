package kin8tengoku

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/gocolly/colly/v2"
	"golang.org/x/net/html"
	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/common/parser"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.MovieProvider = (*KIN8)(nil)

const (
	Name     = "KIN8"
	Priority = 1000
)

const (
	baseURL   = "https://www.kin8tengoku.com/"
	movieURL  = "https://www.kin8tengoku.com/moviepages/%04s/index.html"
	reviewURL = "https://m-template.heyzo.com/snstb/api/review?%s"
)

type KIN8 struct {
	*scraper.Scraper
}

func New() *KIN8 {
	return &KIN8{scraper.NewDefaultScraper(Name, baseURL, Priority, language.Japanese)}
}

func (k8 *KIN8) NormalizeMovieID(id string) string {
	if ss := regexp.MustCompile(`^(?i)(?:kin8[-_])?(\d+)$`).FindStringSubmatch(id); len(ss) == 2 {
		return ss[1]
	}
	return ""
}

func (k8 *KIN8) GetMovieInfoByID(id string) (info *model.MovieInfo, err error) {
	return k8.GetMovieInfoByURL(fmt.Sprintf(movieURL, id))
}

func (k8 *KIN8) ParseMovieIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return path.Base(path.Dir(homepage.Path)), nil
}

func (k8 *KIN8) GetMovieInfoByURL(rawURL string) (info *model.MovieInfo, err error) {
	id, err := k8.ParseMovieIDFromURL(rawURL)
	if err != nil {
		return
	}

	info = &model.MovieInfo{
		ID:            id,
		Number:        fmt.Sprintf("KIN8-%s", id),
		Provider:      k8.Name(),
		Homepage:      rawURL,
		Maker:         "金髪天國",
		Actors:        []string{},
		PreviewImages: []string{},
		Genres:        []string{},
	}

	c := k8.ClonedCollector()

	// Title
	c.OnXML(`//*[@id="sub_main"]/p[@class="sub_title" or @class="sub_title_vip"]`, func(e *colly.XMLElement) {
		info.Title = strings.TrimSpace(e.Text)
	})

	// Title (fallback)
	c.OnXML(`//meta[@name="keywords"]`, func(e *colly.XMLElement) {
		if info.Title != "" {
			return
		}
		info.Title = strings.TrimSpace(e.Attr("content"))
	})

	// Summary
	c.OnXML(`//*[@id="comment"]`, func(e *colly.XMLElement) {
		info.Summary = strings.TrimSpace(e.Text)
	})

	// Fields
	c.OnXML(`//*[@id="detail_box" or @id="detail_box_vip"]//tr`, func(e *colly.XMLElement) {
		switch e.ChildText(`//td[@class="movie_table_td"]`) {
		case "モデル":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `//td[@class="movie_table_td2"]`),
				(*[]string)(&info.Actors))
		case "カテゴリー":
			parser.ParseTexts(htmlquery.FindOne(e.DOM.(*html.Node), `//td[@class="movie_table_td2"]`),
				(*[]string)(&info.Genres))
		case "再生時間":
			info.Runtime = parser.ParseRuntime(e.ChildText(`//td[@class="movie_table_td2"]`))
		case "更新日":
			info.ReleaseDate = parser.ParseDate(e.ChildText(`//td[@class="movie_table_td2"]`))
		}
	})

	// Thumb+Cover+Preview Video
	c.OnXML(`//*[@id="movie"]/script`, func(e *colly.XMLElement) {
		if info.CoverURL != "" {
			return
		}
		if re := regexp.MustCompile(`imgurl\s*=\s*'(.+?)';`); re.MatchString(e.Text) {
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 2 {
				info.CoverURL = e.Request.AbsoluteURL(ss[1])
				info.ThumbURL = info.CoverURL /* use cover as thumb */
			}
		}
		if re := regexp.MustCompile(`samplelimit\s*=\s*(\d+);`); re.MatchString(e.Text) {
			if ss := re.FindStringSubmatch(e.Text); len(ss) == 2 {
				sampleLimit, _ := strconv.Atoi(ss[1])
				videoID, _ := strconv.Atoi(info.ID)
				if sampleLimit > 0 && videoID > 0 {
					// get video urls.
					if re := regexp.MustCompile(`videourl\s*=\s*'(.+?)'`); re.MatchString(e.Text) {
						if ss := re.FindAllStringSubmatch(e.Text, -1); len(ss) == 1 {
							info.PreviewVideoURL = e.Request.AbsoluteURL(ss[0][1])
						} else if len(ss) == 2 {
							if videoID >= sampleLimit {
								info.PreviewVideoURL = e.Request.AbsoluteURL(ss[0][1])
							} else {
								info.PreviewVideoURL = e.Request.AbsoluteURL(ss[1][1])
							}
						}
					}
				}
			}
		}
	})

	// Preview Images
	c.OnXML(`//*[@id="gallery" or @id="gallery_vip"]/div/a`, func(e *colly.XMLElement) {
		if href := e.Attr("href"); href != "" {
			info.PreviewImages = append(info.PreviewImages, e.Request.AbsoluteURL(href))
		}
	})

	// Score
	//$(document).ready(function() {
	//	var d2p_site_id = $('#review_list').data('d2p_site_id');
	//	var movie_seq = $('#review_list').data('movie_seq');
	//
	//	var domain_app_server_public = 'm-template.heyzo.com';
	//
	//	$.getJSON('//' + domain_app_server_public + '/snstb/api/review?callback=?', {
	//		site_id: d2p_site_id,
	//		movie_seq: movie_seq,
	//		json: 1
	//	}, function(json) {
	//		//debugger; //check what you get in json object when you use debugger.
	//		var html = "";
	//		var product_id = "";
	//		$.each(json, function(i, review) {
	//			// index, each data
	//			if (product_id == '') {
	//				product_id = review.product_id;
	//			}
	//			html += "<div class='rev-box guest'>";
	//			html += "<div class='usr-rev-area'>";
	//			html += "<h5><span class='rev_star" + review.user_rating + "'></span></h5>";
	//			html += "<p>" + review.user_comment + "</p>";
	//			html += "</div><!-- .usr-rev-area -->";
	//			html += "<div class='usr-foot-area'><p class='px11'>by <span>" + review.profile_name + "</span>投稿日 : " + review.created_jp + "</p></div>";
	//			html += "</div><!-- .rev-box -->";
	//		});
	//		if (product_id != '') {
	//			html += "<div class='usr-button-see-more seeall'><a target='_blank' href='https://sns.d2pass.com/product/movies/" + product_id + "'>もっと見る</a></div>";
	//		}
	//		$("#review_overflow").html(html);
	//	});
	//})
	c.OnXML(`//*[@id="review_list"]`, func(e *colly.XMLElement) {
		siteId := e.Attr("data-d2p_site_id")
		movieSeq := e.Attr("data-movie_seq")
		jqTimestamp := strconv.Itoa(int(time.Now().UnixMilli()))
		d := c.Clone()
		d.OnResponse(func(r *colly.Response) {
			if ss := regexp.MustCompile(`(?s)\w+\((.+?)\);`).FindSubmatch(r.Body); len(ss) == 2 {
				var data []struct {
					UserRating float64 `json:"user_rating"`
				}
				if json.Unmarshal(ss[1], &data) == nil {
					total := 0.0
					for _, i := range data {
						total += i.UserRating
					}
					info.Score = total / float64(len(data))
				}
			}
		})
		q := &url.Values{}
		q.Add("callback", fmt.Sprintf("jQuery_%s", jqTimestamp))
		q.Add("site_id", siteId)
		q.Add("movie_seq", movieSeq)
		q.Add("json", "1")
		q.Add("_", jqTimestamp)
		d.Visit(fmt.Sprintf(reviewURL, q.Encode()))
	})

	err = c.Visit(info.Homepage)
	return
}

func init() {
	provider.Register(Name, New)
}
