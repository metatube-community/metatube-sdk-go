package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/model"
	"github.com/metatube-community/metatube-sdk-go/plugin"
	// register plugins
	_ "github.com/metatube-community/metatube-sdk-go/plugin/actorinfo"
	_ "github.com/metatube-community/metatube-sdk-go/plugin/movieinfo"
)

type infoType uint8

const (
	actorInfoType infoType = iota
	movieInfoType
)

type infoUri struct {
	Provider string `uri:"provider" binding:"required"`
	ID       string `uri:"id" binding:"required"`
}

type infoQuery struct {
	Lazy    bool   `form:"lazy"`
	Plugins string `form:"plugins"`
}

func getInfo(app *engine.Engine, typ infoType) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := &infoUri{}
		if err := c.ShouldBindUri(uri); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}
		query := &infoQuery{
			Lazy: true, // enable lazy by default.
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		// parse plugins.
		plugins, err := parsePlugins(typ, query.Plugins)
		if err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		var info any
		switch typ {
		case actorInfoType:
			info, err = app.GetActorInfoByProviderID(uri.Provider, uri.ID, query.Lazy)
		case movieInfoType:
			info, err = app.GetMovieInfoByProviderID(uri.Provider, uri.ID, query.Lazy)
		default:
			panic("invalid info/metadata type")
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		// apply plugins.
		for _, apply := range plugins {
			switch typ {
			case actorInfoType:
				apply.(plugin.ActorInfoPlugin)(app, info.(*model.ActorInfo))
			case movieInfoType:
				apply.(plugin.MovieInfoPlugin)(app, info.(*model.MovieInfo))
			}
		}

		c.JSON(http.StatusOK, &responseMessage{Data: info})
	}
}

func parsePlugins(typ infoType, raw string) (plugins []any, err error) {
	for _, name := range strings.Split(raw, ",") {
		if name = strings.TrimSpace(name); name != "" {
			var (
				pl any
				ok bool
			)
			switch typ {
			case actorInfoType:
				pl, ok = plugin.LookupActorInfoPlugin(name)
			case movieInfoType:
				pl, ok = plugin.LookupMovieInfoPlugin(name)
			}
			if !ok {
				err = fmt.Errorf("unsupported plugin: %s", name)
				return
			}
			plugins = append(plugins, pl)
		}
	}
	return
}
