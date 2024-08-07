package engine

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/database"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

const (
	DefaultEngineName     = "metatube"
	DefaultRequestTimeout = time.Minute
)

// Special environment prefixes for setting provider priorities.
const (
	ActorProviderPriorityEnvPrefix = "MT_ACTOR_PROVIDER_PRIORITY_"
	MovieProviderPriorityEnvPrefix = "MT_MOVIE_PROVIDER_PRIORITY_"
)

type Engine struct {
	db      *gorm.DB
	name    string
	timeout time.Duration
	fetcher *fetch.Fetcher
	// Engine Logger
	logger *zap.SugaredLogger
	// Name:Provider Map
	actorProviders map[string]mt.ActorProvider
	movieProviders map[string]mt.MovieProvider
	// Host:Providers Map
	actorHostProviders map[string][]mt.ActorProvider
	movieHostProviders map[string][]mt.MovieProvider
}

func New(db *gorm.DB, opts ...Option) *Engine {
	engine := &Engine{
		db:      db,
		name:    DefaultEngineName,
		timeout: DefaultRequestTimeout,
	}
	// apply options
	for _, opt := range opts {
		opt(engine)
	}
	return engine.init()
}

func Default() *Engine {
	db, _ := database.Open(&database.Config{
		DSN:                  "",
		DisableAutomaticPing: true,
	})
	engine := New(db)
	defer engine.AutoMigrate(true)
	return engine
}

func (e *Engine) init() *Engine {
	logger, _ := zap.NewProduction()
	e.logger = logger.Sugar()
	e.fetcher = fetch.Default(&fetch.Config{
		Timeout: e.timeout,
	})
	e.initActorProviders(e.timeout)
	e.initMovieProviders(e.timeout)
	e.initAllProviderPriorities()
	return e
}

func (e *Engine) initAllProviderPriorities() {
	for _, env := range os.Environ() {
		key, value, _ := strings.Cut(strings.ToUpper(env), "=")
		switch {
		case strings.HasPrefix(key, ActorProviderPriorityEnvPrefix):
			name := key[len(ActorProviderPriorityEnvPrefix):]
			prio, _ := strconv.ParseInt(value, 0, 64)
			if prio == 0 {
				delete(e.actorProviders, name)
				continue
			}
			if provider, ok := e.actorProviders[name]; ok {
				provider.SetPriority(prio)
			}
		case strings.HasPrefix(key, MovieProviderPriorityEnvPrefix):
			name := key[len(MovieProviderPriorityEnvPrefix):]
			prio, _ := strconv.ParseInt(value, 0, 64)
			if prio == 0 {
				delete(e.movieProviders, name)
				continue
			}
			if provider, ok := e.movieProviders[name]; ok {
				provider.SetPriority(prio)
			}
		}
	}
}

// initActorProviders initializes actor providers.
func (e *Engine) initActorProviders(timeout time.Duration) {
	{ // init
		e.actorProviders = make(map[string]mt.ActorProvider)
		e.actorHostProviders = make(map[string][]mt.ActorProvider)
	}
	mt.RangeActorFactory(func(name string, factory mt.ActorFactory) {
		provider := factory()
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		// Add actor provider by name.
		e.actorProviders[strings.ToUpper(name)] = provider
		// Add actor provider by host.
		host := provider.URL().Hostname()
		e.actorHostProviders[host] = append(e.actorHostProviders[host], provider)
	})
}

// initMovieProviders initializes movie providers.
func (e *Engine) initMovieProviders(timeout time.Duration) {
	{ // init
		e.movieProviders = make(map[string]mt.MovieProvider)
		e.movieHostProviders = make(map[string][]mt.MovieProvider)
	}
	mt.RangeMovieFactory(func(name string, factory mt.MovieFactory) {
		provider := factory()
		if s, ok := provider.(mt.RequestTimeoutSetter); ok {
			s.SetRequestTimeout(timeout)
		}
		// Add movie provider by name.
		e.movieProviders[strings.ToUpper(name)] = provider
		// Add movie provider by host.
		host := provider.URL().Hostname()
		e.movieHostProviders[host] = append(e.movieHostProviders[host], provider)
	})
}

func (e *Engine) IsActorProvider(name string) (ok bool) {
	_, ok = e.actorProviders[strings.ToUpper(name)]
	return
}

func (e *Engine) GetActorProviders() map[string]mt.ActorProvider {
	return e.actorProviders
}

func (e *Engine) GetActorProviderByURL(rawURL string) (mt.ActorProvider, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	for _, p := range e.actorHostProviders[u.Hostname()] {
		if strings.HasPrefix(u.Path, p.URL().Path) {
			return p, nil
		}
	}
	return nil, mt.ErrProviderNotFound
}

func (e *Engine) GetActorProviderByName(name string) (mt.ActorProvider, error) {
	provider, ok := e.actorProviders[strings.ToUpper(name)]
	if !ok {
		return nil, mt.ErrProviderNotFound
	}
	return provider, nil
}

func (e *Engine) MustGetActorProviderByName(name string) mt.ActorProvider {
	provider, err := e.GetActorProviderByName(name)
	if err != nil {
		panic(err)
	}
	return provider
}

func (e *Engine) IsMovieProvider(name string) (ok bool) {
	_, ok = e.movieProviders[strings.ToUpper(name)]
	return
}

func (e *Engine) GetMovieProviders() map[string]mt.MovieProvider {
	return e.movieProviders
}

func (e *Engine) GetMovieProviderByURL(rawURL string) (mt.MovieProvider, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	for _, p := range e.movieHostProviders[u.Hostname()] {
		if strings.HasPrefix(u.Path, p.URL().Path) {
			return p, nil
		}
	}
	return nil, mt.ErrProviderNotFound
}

func (e *Engine) GetMovieProviderByName(name string) (mt.MovieProvider, error) {
	provider, ok := e.movieProviders[strings.ToUpper(name)]
	if !ok {
		return nil, mt.ErrProviderNotFound
	}
	return provider, nil
}

func (e *Engine) MustGetMovieProviderByName(name string) mt.MovieProvider {
	provider, err := e.GetMovieProviderByName(name)
	if err != nil {
		panic(err)
	}
	return provider
}

func (e *Engine) AutoMigrate(v bool) error {
	if !v {
		return nil
	}
	// Create Case-Insensitive Collation for Postgres.
	if e.db.Config.Dialector.Name() == database.Postgres {
		e.db.Exec(`CREATE COLLATION IF NOT EXISTS NOCASE (
		provider = icu,
		locale = 'und-u-ks-level2',
		deterministic = FALSE)`)
	}
	return e.db.AutoMigrate(
		&model.MovieInfo{},
		&model.ActorInfo{},
		&model.MovieReviewInfo{},
	)
}

// Fetch fetches content from url. If provider is nil, the
// default fetcher will be used.
func (e *Engine) Fetch(url string, provider mt.Provider) (*http.Response, error) {
	// Provider which implements Fetcher interface should be
	// used to fetch all its corresponding resources.
	if fetcher, ok := provider.(mt.Fetcher); ok {
		return fetcher.Fetch(url)
	}
	return e.fetcher.Fetch(url)
}

// String returns the name of the Engine instance.
func (e *Engine) String() string {
	return e.name
}
