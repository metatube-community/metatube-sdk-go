package route

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/errors"
	"github.com/javtube/javtube-sdk-go/httputil"
	V "github.com/javtube/javtube-sdk-go/internal/constant"
	"github.com/javtube/javtube-sdk-go/route/validator"
)

func New(app *engine.Engine, v validator.Validator) *gin.Engine {
	r := gin.New()
	{
		// register middleware
		r.Use(logger(), recovery())
		// fallback behavior
		r.NoRoute(notFound())
		r.NoMethod(notAllowed())
		// index page
		r.GET("/", index())
	}

	// redirection middleware
	r.Use(redirect(app))

	api := r.Group("/api")
	api.Use(authentication(v))
	{
		// info/metadata
		api.GET("/actor", getInfo(app, actorInfoType))
		api.GET("/movie", getInfo(app, movieInfoType))

		// translate
		api.GET("/translate", getTranslate(defaultMaxRPS))

		// search
		search := api.Group("/search")
		search.GET("/actor", getSearchResults(app, actorSearchType))
		search.GET("/movie", getSearchResults(app, movieSearchType))
	}

	img := r.Group("/image")
	{
		img.GET("/primary", getImage(app, primaryImageType))
		img.GET("/thumb", getImage(app, thumbImageType))
		img.GET("/backdrop", getImage(app, backdropImageType))
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

func index() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, &responseMessage{
			Success: true,
			Data: gin.H{
				"version":    V.Version,
				"git-commit": V.GitCommit,
			},
		})
	}
}

func abortWithError(c *gin.Context, err error) {
	var code = http.StatusInternalServerError
	if e, ok := err.(*errors.HTTPError); ok {
		code = e.StatusCode()
	} else if c := httputil.StatusCode(err.Error()); c != 0 {
		code = c
	}
	abortWithStatusMessage(c, code, err)
}

func abortWithStatusMessage(c *gin.Context, code int, message any) {
	switch m := message.(type) {
	case string:
		// pass
	case error:
		message = m.Error()
	case fmt.Stringer:
		message = m.String()
	default:
		message = fmt.Sprintf("%#v", m)
	}
	c.AbortWithStatusJSON(code, &responseMessage{
		Success: false,
		Error: &errors.HTTPError{
			Code:    code,
			Message: message.(string),
		},
	})
}

type responseMessage struct {
	Success bool  `json:"success"`
	Data    any   `json:"data,omitempty"`
	Error   error `json:"error,omitempty"`
}
