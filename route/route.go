package route

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/errors"
	V "github.com/javtube/javtube-sdk-go/internal/constant"
	"github.com/javtube/javtube-sdk-go/route/auth"
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

	v1 := r.Group("/v1")
	v1.Use(authentication(v))
	{
		// translate
		v1.GET("/translate", getTranslate(defaultMaxRPS))

		actors := v1.Group("/actors")
		{
			actors.GET("/:provider/:id", getInfo(app, actorInfoType))
			actors.GET("/search", getSearch(app, actorSearchType))
		}

		movies := v1.Group("/movies")
		{
			movies.GET("/:provider/:id", getInfo(app, movieInfoType))
			movies.GET("/search", getSearch(app, movieSearchType))
		}
	}

	images := r.Group("/images")
	{
		images.GET("/primary/:provider/:id", getImage(app, primaryImageType))
		images.GET("/thumb/:provider/:id", getImage(app, thumbImageType))
		images.GET("/backdrop/:provider/:id", getImage(app, backdropImageType))
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
				"app":     "javtube",
				"commit":  V.GitCommit,
				"version": V.Version,
			},
		})
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
