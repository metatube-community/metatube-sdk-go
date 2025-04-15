package gfriends

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"slices"
	"time"

	"golang.org/x/text/language"

	"github.com/metatube-community/metatube-sdk-go/collections"
	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/singledo"
	"github.com/metatube-community/metatube-sdk-go/provider"
	"github.com/metatube-community/metatube-sdk-go/provider/internal/scraper"
)

var _ provider.ActorImageProvider = (*Gfriends)(nil)

const (
	Name     = "Gfriends"
	Priority = 1000 - 1
)

const (
	baseURL    = "https://github.com/gfriends/gfriends"
	contentURL = "https://raw.githubusercontent.com/gfriends/gfriends/master/Content/%s/%s"
	jsonURL    = "https://raw.githubusercontent.com/gfriends/gfriends/master/Filetree.json"
)

type Gfriends struct {
	*scraper.Scraper
}

func New() *Gfriends {
	return &Gfriends{scraper.NewDefaultScraper(
		Name, baseURL, Priority,
		language.Japanese,
		scraper.WithDisableCookies(),
	)}
}

func (gf *Gfriends) GetActorImagesByName(name string) ([]string, error) {
	images, err := _fileTree.query(name)
	if err != nil {
		return nil, err
	}
	return images, nil
}

var (
	_fileTree = newFileTree(2 * time.Hour)
	_fetcher  = fetch.Default(nil)
)

type fileTree struct {
	single *singledo.Single

	// `Content`
	Content *collections.OrderedMap[string, *collections.OrderedMap[string, string]] `json:"Content"`

	// `Information`
	//Information struct {
	//	TotalNum  int     `json:"TotalNum"`
	//	TotalSize int     `json:"TotalSize"`
	//	Timestamp float64 `json:"Timestamp"`
	//} `json:"Information"`
}

func newFileTree(wait time.Duration) *fileTree {
	return &fileTree{
		single:  singledo.NewSingle(wait),
		Content: collections.NewOrderedMap[string, *collections.OrderedMap[string, string]](),
	}
}

func (ft *fileTree) query(s string) (images []string, err error) {
	// update
	ft.single.Do(func() (any, error) {
		err = ft.update()
		return nil, nil
	})
	// query
	for co, am := range ft.Content.Iterator() {
		for n, p := range am.Iterator() {
			if n[:len(n)-len(path.Ext(n))] == s /* exact match */ {
				if u, e := url.Parse(fmt.Sprintf(contentURL, co, p)); e == nil {
					images = append(images, u.String())
				}
			}
		}
	}
	slices.Reverse(images) // descending
	return
}

func (ft *fileTree) update() error {
	resp, err := _fetcher.Fetch(jsonURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(ft)
}

func init() {
	provider.Register(Name, New)
}
