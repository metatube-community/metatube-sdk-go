package movieinfo

import (
	"fmt"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/plugin"
)

func WithRealActorNames() plugin.MovieInfoPlugin {
	const (
		avBase    = "AVBASE"
		namespace = "real_actor_names"
	)
	var providers = map[string]struct{}{
		"DUGA":   {},
		"FANZA":  {},
		"GETCHU": {},
		"MGS":    {},
		"PCOLLE": {},
	}
	return func(app *engine.Engine, info *model.MovieInfo) {
		if _, ok := providers[strings.ToUpper(info.Provider)]; !ok {
			app.Logger().
				Named(namespace).
				Infof("skip unsupported provider: %s", info.Provider)
			return
		}

		// search from DB first
		results, err := app.SearchMovieFromDB(info.Number, avBase, false)
		if err != nil {
			app.Logger().
				Named(namespace).
				Warnf("search movie from DB %s: %v", info.Number, err)
			// ignore error and proceed
		}

		found := false
		if len(results) != 1 {
			app.Logger().
				Named(namespace).
				Infof("zero/multiple movie(s) found from DB: %s", info.ID)
		} else {
			app.Logger().
				Named(namespace).
				Infof("movie found from DB: %s", info.ID)
			found = true
			goto results
		}

		// search from AVBASE.
		results, err = app.SearchMovie(info.ID, avBase, false)
		if err != nil {
			app.Logger().
				Named(namespace).
				Warnf("search movie: %v", err)
			return
		}

	results:
		switch {
		case len(results) == 0:
			app.Logger().
				Named(namespace).
				Warnf("movie not found: %s", info.ID)
		case len(results) > 1:
			app.Logger().
				Named(namespace).
				Warnf("multiple movies found: %s", info.ID)
		default:
			if !found { // make it store to the DB, ignore errors
				if _, err := app.GetMovieInfoByProviderID(
					results[0].Provider,
					results[0].ID,
					false,
				); err != nil {
					app.Logger().
						Named(namespace).
						Warnf("fetch movie info %s: %v", info.ID, err)
				}
			}
			info.Actors = results[0].Actors
		}
		return
	}
}

func WithActorDetails() plugin.MovieInfoPlugin {
	return func(app *engine.Engine, info *model.MovieInfo) {
		getActorInfo := func(actor string) (ai *model.ActorInfo, err error) {
			app.SearchActorFromDB(actor)
			results, err := app.SearchActorAll(actor, false)
			if err != nil || len(results) == 0 {
				err = fmt.Errorf("actor %s not found: %v", actor, err)
				return
			}
			return app.GetActorInfoByProviderID(results[0].Provider, results[0].ID, false)
		}

		var actorDetails []*model.ActorInfo
		for _, actor := range info.Actors {
			ai, err := getActorInfo(actor)
			if err != nil {
				app.Logger().
					Named("actor_details").
					Warnf("get actor info for %s: %v", actor, err)
				// ignore error and continue
				continue
			}
			actorDetails = append(actorDetails, ai)
		}
		info.ActorDetails = actorDetails
	}
}

func WithActorDetailsGFriends() plugin.MovieInfoPlugin {
	const gFriends = "GFriends"
	return func(app *engine.Engine, info *model.MovieInfo) {
		var actorDetails []*model.ActorInfo
		for _, actor := range info.Actors {
			ai, err := app.GetActorInfoByProviderID(gFriends, actor, false)
			if err != nil {
				app.Logger().
					Named("gfriends").
					Warnf("get actor info for %s: %v", actor, err)
				return
			}
			actorDetails = append(actorDetails, ai)
		}
		info.ActorDetails = actorDetails
	}
}

func init() {
	plugin.RegisterMovieInfoPlugin("real_actor_names", WithRealActorNames())
	plugin.RegisterMovieInfoPlugin("actor_details", WithActorDetails())
	plugin.RegisterMovieInfoPlugin("gfriends", WithActorDetailsGFriends())
}
