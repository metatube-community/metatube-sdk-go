package gfriends

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/iancoleman/orderedmap"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/common/reverse"
	"github.com/metatube-community/metatube-sdk-go/common/singledo"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/provider"
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
	baseURL    = "https://github.com/gfriends/gfriends"
	contentURL = "https://raw.githubusercontent.com/gfriends/gfriends/master/Content/%s/%s"
	jsonURL    = "https://raw.githubusercontent.com/gfriends/gfriends/master/Filetree.json"
)

var (
	_baseURL = mustParse(baseURL)
	_fetcher = fetch.Default(nil)
)

type GFriends struct{}

func New() *GFriends { return &GFriends{} }

func (gf *GFriends) Name() string { return Name }

func (gf *GFriends) Priority() int { return Priority }

func (gf *GFriends) URL() *url.URL { return _baseURL }

func (gf *GFriends) NormalizeID(id string) string { return id /* AS IS */ }

func (gf *GFriends) GetActorInfoByID(id string) (*model.ActorInfo, error) {
	images, err := defaultFileTree.query(id)
	if len(images) == 0 {
		if err != nil {
			return nil, err
		}
		return nil, provider.ErrInfoNotFound
	}
	return &model.ActorInfo{
		ID:       id,
		Name:     id,
		Provider: gf.Name(),
		Homepage: gf.formatURL(id),
		Aliases:  []string{},
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
	single *singledo.Single

	// `Content`
	Content *orderedmap.OrderedMap `json:"Content"`

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
		Content: orderedmap.New(),
	}
}

func (ft *fileTree) query(s string) (images []string, err error) {
	// update
	ft.single.Do(func() (any, error) {
		err = ft.update()
		return nil, nil
	})
	// query
	for _, c := range ft.Content.Keys() {
		if o, ok := ft.Content.Get(c); ok {
			am := o.(orderedmap.OrderedMap)
			for _, n := range am.Keys() {
				if n[:len(n)-len(path.Ext(n))] == s /* exact match */ {
					p, _ := am.Get(n)
					if u, e := url.Parse(fmt.Sprintf(contentURL, c, p.(string))); e == nil {
						images = append(images, u.String())
					}
				}
			}
		}
	}
	reverse.Slice(images) // descending
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

func mustParse(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

func init() {
	provider.RegisterActorFactory(Name, New)
}
