package movieinfo

import (
	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/plugin"
)

func WithReviews() plugin.MovieInfoPlugin {
	return func(app *engine.Engine, info *model.MovieInfo) {
		reviews, err := app.GetMovieReviewsByProviderURL(info.Provider, info.Homepage, true)
		if err != nil {
			app.Logger().
				Named("reviews").
				Warnf("get reviews for <%s:%s>: %v", info.Provider, info.ID, err)
			return
		}
		info.Reviews = reviews.Reviews.Data()
	}
}

func init() {
	plugin.RegisterMovieInfoPlugin("reviews", WithReviews())
}
