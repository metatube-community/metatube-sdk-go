package gfriends

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/javtube/javtube-sdk-go/model"
	"github.com/javtube/javtube-sdk-go/provider"
)

var (
	_ provider.ActorProvider = (*GFriends)(nil)
	_ provider.ActorSearcher = (*GFriends)(nil)
)

const (
	name     = "gfriends"
	priority = 9
)

const (
	baseURL    = "https://github.com/xinxin8816/gfriends"
	contentURL = "https://raw.githubusercontent.com/xinxin8816/gfriends/master/Content/%s"
	jsonURL    = "https://raw.githubusercontent.com/xinxin8816/gfriends/master/Filetree.json"
)

type GFriends struct {
	fileTree *fileTree
}

func New() *GFriends {
	return &GFriends{
		fileTree: newFileTree(time.Hour),
	}
}

func (gf *GFriends) Name() string {
	return name
}

func (gf *GFriends) Priority() int {
	return priority
}

func (gf *GFriends) NormalizeID(id string) string { return id /* AS IS */ }

func (gf *GFriends) GetActorInfoByID(id string) (*model.ActorInfo, error) {
	images, err := gf.fileTree.query(id)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, provider.ErrNotFound
	}
	return &model.ActorInfo{
		ID:       id,
		Name:     id,
		Provider: gf.Name(),
		Homepage: baseURL,
		Aliases:  []string{},
		Images:   images,
	}, nil
}

func (gf *GFriends) GetActorInfoByURL(string) (*model.ActorInfo, error) {
	return nil, provider.ErrNotSupported
}

func (gf *GFriends) SearchActor(keyword string) (results []*model.ActorSearchResult, err error) {
	var info *model.ActorInfo
	if info, err = gf.GetActorInfoByID(keyword); err == nil && info.Valid() {
		results = []*model.ActorSearchResult{info.ToSearchResult()}
	}
	return
}

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
	resp, err := http.Get(jsonURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return json.NewDecoder(resp.Body).Decode(ft)
	}
	return errors.New(http.StatusText(resp.StatusCode))
}

func reverse[T any](array []T) []T {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}
	return array
}

func init() {
	provider.RegisterActorFactory(name, New)
}
