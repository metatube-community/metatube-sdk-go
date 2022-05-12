package engine

import (
	"time"

	"github.com/javtube/javtube-sdk-go/model"
	javtube "github.com/javtube/javtube-sdk-go/provider"
	"gorm.io/gorm"
)

type Engine struct {
	db             *gorm.DB
	movieProviders map[string]javtube.MovieProvider
	actorProviders map[string]javtube.ActorProvider
}

func New(db *gorm.DB, timeout time.Duration) (engine *Engine) {
	engine = &Engine{
		db:             db,
		movieProviders: make(map[string]javtube.MovieProvider),
		actorProviders: make(map[string]javtube.ActorProvider),
	}
	// Initialize movie providers.
	javtube.RangeMovieFactory(func(name string, factory javtube.MovieFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok && timeout > 0 {
			s.SetRequestTimeout(timeout)
		}
		engine.movieProviders[name] = provider
	})
	// Initialize actor providers.
	javtube.RangeActorFactory(func(name string, factory javtube.ActorFactory) {
		provider := factory()
		if s, ok := provider.(javtube.RequestTimeoutSetter); ok && timeout > 0 {
			s.SetRequestTimeout(timeout)
		}
		engine.actorProviders[name] = factory()
	})
	return
}

func (e *Engine) AutoMigrate() error {
	return e.db.AutoMigrate(
		&model.MovieInfo{},
		&model.ActorInfo{})
}

func (e *Engine) Download() {

}
