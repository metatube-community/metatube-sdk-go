package route

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/engine"
	"github.com/javtube/javtube-sdk-go/errors"
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
	}

	// redirection middleware
	r.Use(redirect(app))

	r.GET("/", getIndex())
	r.GET("/version", getVersion())

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
		search.GET("/actor", getSearch(app, actorSearchType))
		search.GET("/movie", getSearch(app, movieSearchType))
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

func getIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, &responseMessage{
			Success: true,
			Data:    gin.H{"hello": "javtube"},
		})
	}
}

func getVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, &responseMessage{
			Success: true,
			Data: &struct {
				Version   string `json:"version"`
				GitCommit string `json:"git_commit"`
			}{
				Version:   V.Version,
				GitCommit: V.GitCommit,
			},
		})
	}
}

func abortWithError(c *gin.Context, err error) {
	if e, ok := err.(*errors.HTTPError); ok {
		c.AbortWithStatusJSON(e.Code, &responseMessage{
			Success: false,
			Error:   e,
		})
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
		Success: false,
		Error:   errors.New(code, fmt.Sprintf("%v", message)),
	})
}

type responseMessage struct {
	Success bool  `json:"success"`
	Data    any   `json:"data,omitempty"`
	Error   error `json:"error,omitempty"`
}
