package route

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/errors"
	V "github.com/metatube-community/metatube-sdk-go/internal/version"
	"github.com/metatube-community/metatube-sdk-go/route/auth"
)

func New(app *engine.Engine, v auth.Validator) *gin.Engine {
	r := gin.New()
	{
		// register middleware
		r.Use(logger(), recovery())
		// fallback behavior
		r.NoRoute(notFound())
		r.NoMethod(notAllowed())
	}

	// redirection middleware
	r.Use(redirect(app))

	// index page
	r.GET("/", getIndex())

	public := r.Group("/v1")
	{
		public.GET("/translate", getTranslate())

		public.GET("/providers", getProviders(app))

		images := public.Group("/images")
		{
			images.GET("/primary/:provider/:id", getImage(app, primaryImageType))
			images.GET("/thumb/:provider/:id", getImage(app, thumbImageType))
			images.GET("/backdrop/:provider/:id", getImage(app, backdropImageType))
		}
	}

	private := r.Group("/v1", authentication(v))
	{
		actors := private.Group("/actors")
		{
			actors.GET("/:provider/:id", getInfo(app, actorInfoType))
			actors.GET("/search", getSearch(app, actorSearchType))
		}

		movies := private.Group("/movies")
		{
			movies.GET("/:provider/:id", getInfo(app, movieInfoType))
			movies.GET("/search", getSearch(app, movieSearchType))
		}
	}

	return r
}

func logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{})
}

func recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		abortWithStatusMessage(c, http.StatusInternalServerError, err)
	})
}

func notFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		abortWithStatusMessage(c, http.StatusNotFound,
			http.StatusText(http.StatusNotFound))
	}
}

func notAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		abortWithStatusMessage(c, http.StatusMethodNotAllowed,
			http.StatusText(http.StatusMethodNotAllowed))
	}
}

func getIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, &responseMessage{
			Data: gin.H{
				"app":     "metatube",
				"commit":  V.GitCommit,
				"version": V.Version,
			},
		})
	}
}

func getProviders(app *engine.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			ActorProviders map[string]string `json:"actor_providers"`
			MovieProviders map[string]string `json:"movie_providers"`
		}{
			ActorProviders: make(map[string]string),
			MovieProviders: make(map[string]string),
		}
		for _, provider := range app.GetActorProviders() {
			data.ActorProviders[provider.Name()] = provider.URL().String()
		}
		for _, provider := range app.GetMovieProviders() {
			data.MovieProviders[provider.Name()] = provider.URL().String()
		}
		c.JSON(http.StatusOK, &responseMessage{Data: data})
	}
}

func abortWithError(c *gin.Context, err error) {
	if e, ok := err.(*errors.HTTPError); ok {
		c.AbortWithStatusJSON(e.Code, &responseMessage{Error: e})
		return
	}
	code := http.StatusInternalServerError
	if c := errors.StatusCode(err); c != 0 {
		code = c
	}
	abortWithStatusMessage(c, code, err)
}

func abortWithStatusMessage(c *gin.Context, code int, message any) {
	c.AbortWithStatusJSON(code, &responseMessage{
		Error: errors.New(code, fmt.Sprintf("%v", message)),
	})
}

type responseMessage struct {
	Data  any   `json:"data,omitempty"`
	Error error `json:"error,omitempty"`
}
