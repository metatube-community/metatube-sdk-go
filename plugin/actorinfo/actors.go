package actorinfo

import (
	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/plugin"
)

func WithDemo() plugin.ActorInfoPlugin {
	return func(app *engine.Engine, info *model.ActorInfo) {
		app.Logger().
			Named("demo").
			Infof("actor homepage: %s", info.Homepage)
	}
}

func init() {
	plugin.RegisterActorInfoPlugin("demo", WithDemo())
}
