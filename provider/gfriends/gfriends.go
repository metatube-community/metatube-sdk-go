package gfriends

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/iancoleman/orderedmap"

	"github.com/javtube/javtube-sdk-go/common/fetch"
	"github.com/javtube/javtube-sdk-go/common/parser"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
)

var (
	_ provider.ActorProvider = (*GFriends)(nil)
	_ provider.ActorSearcher = (*GFriends)(nil)
)

const (
	Name     = "GFriends"
	Priority = 1000 - 1
)

const gFriendsID = "gfriends-id"

const (
	baseURL    = "https://github.com/xinxin8816/gfriends"
	contentURL = "https://raw.githubusercontent.com/xinxin8816/gfriends/master/Content/%s"
	jsonURL    = "https://raw.githubusercontent.com/xinxin8816/gfriends/master/Filetree.json"
)

var (
	_baseURL, _ = url.Parse(baseURL)
	_fetcher    = fetch.Default(nil)
)

type GFriends struct{}

func New() *GFriends { return &GFriends{} }

func (gf *GFriends) Name() string { return Name }

func (gf *GFriends) Priority() int { return Priority }

func (gf *GFriends) URL() *url.URL { return _baseURL }

func (gf *GFriends) NormalizeID(id string) string { return id /* AS IS */ }

func (gf *GFriends) GetActorInfoByID(id string) (*model.ActorInfo, error) {
	names := parser.ParseActorNames(id)
	if len(names) == 0 {
		return nil, provider.ErrInvalidID
	}
	// use as ordered set.
	set := orderedmap.New()
	for _, name := range names {
		images, err := defaultFileTree.query(name)
		if err != nil &&
			// might be an update issue caused image not found.
			len(images) == 0 {
			return nil, err
		}
		for _, image := range images {
			set.Set(image, struct{}{})
		}
	}
	images := set.Keys()
	if len(images) == 0 {
		return nil, provider.ErrInfoNotFound
	}
	// NOTE: names length must > 1.
	aliases := make([]string, 0, len(names)-1)
	if len(names) > 1 {
		aliases = append(aliases, names[1:]...)
	}
	return &model.ActorInfo{
		ID:       id,
		Name:     names[0], /* simply pick the first actor name */
		Provider: gf.Name(),
		Homepage: gf.formatURL(id),
		Aliases:  aliases,
		Images:   images,
	}, nil
}

func (gf *GFriends) formatURL(id string) string {
	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set(gFriendsID, id)
	u.RawQuery = q.Encode()
	return u.String()
}

func (gf *GFriends) ParseIDFromURL(rawURL string) (string, error) {
	homepage, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return homepage.Query().Get(gFriendsID), nil
}

func (gf *GFriends) GetActorInfoByURL(u string) (*model.ActorInfo, error) {
	id, err := gf.ParseIDFromURL(u)
	if err != nil {
		return nil, err
	}
	return gf.GetActorInfoByID(id)
}

func (gf *GFriends) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	var info *model.ActorInfo
	if info, err = gf.GetActorInfoByID(keyword); err == nil && info.Valid() {
		results = []*model.ActorSearchResult{info.ToSearchResult()}
	}
	return
}

var defaultFileTree = newFileTree(2 * time.Hour)

type fileTree struct {
	mu      sync.RWMutex
	last    time.Time
	timeout time.Duration
	Content *orderedmap.OrderedMap
}

func newFileTree(timeout time.Duration) *fileTree {
	return &fileTree{
		timeout: timeout,
		Content: orderedmap.New(),
	}
}

func (ft *fileTree) query(s string) (images []string, err error) {
	// update
	ft.mu.Lock()
	if ft.last.Add(ft.timeout).Before(time.Now()) {
		if err = ft.update(); err == nil {
			ft.last = time.Now()
		}
	}
	ft.mu.Unlock()
	// query
	ft.mu.RLock()
	defer ft.mu.RUnlock()
	for _, com := range ft.Content.Keys() {
		if o, ok := ft.Content.Get(com); ok {
			am := o.(orderedmap.OrderedMap)
			for _, n := range am.Keys() {
				if n[:len(n)-len(path.Ext(n))] == s /* exact match */ {
					p, _ := am.Get(n)
					images = append(images, fmt.Sprintf(contentURL,
						path.Join(url.PathEscape(com), url.PathEscape(p.(string)))))
				}
			}
		}
	}
	reverse(images) // descending
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

func reverse[T any](array []T) []T {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}
	return array
}

func init() {
	provider.RegisterActorFactory(Name, New)
}
