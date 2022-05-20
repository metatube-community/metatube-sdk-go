package route

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javtube/javtube-sdk-go/engine"
)

func New(app *engine.Engine) *gin.Engine {
	r := gin.New()
	{
		r.Use(logger(), recovery())
		r.NoRoute(notFound())
		r.NoMethod(notAllowed())
	}

	api := r.Group("/api")
	api.Use( /* AUTH */)
	{
		// info/metadata
		api.GET("/actor", getInfo(app, actorInfoType))
		api.GET("/movie", getInfo(app, movieInfoType))

		// search
		search := api.Group("/search")
		search.GET("/actor", getSearchResult(app, actorSearchType))
		search.GET("/movie", getSearchResult(app, movieSearchType))
	}

	img := r.Group("/image")
	{
		img.GET("/primary", GetImage(app, primaryImageType))
		img.GET("/thumb", GetImage(app, thumbImageType))
		img.GET("/backdrop", GetImage(app, backdropImageType))
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

func abortWithStatusMessage(c *gin.Context, code int, message any) {
	switch m := message.(type) {
	case error:
		message = m.Error()
	case fmt.Stringer:
		message = m.String()
	default:
		// skip
	}
	c.AbortWithStatusJSON(code, gin.H{
		"status":  false,
		"message": message,
	})
}
