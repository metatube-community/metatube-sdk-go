package route

import (
	"time"

	"github.com/gin-gonic/gin"
	cachecontrol "go.eigsys.de/gin-cachecontrol/v2"
)

func cacheControl(duration time.Duration) gin.HandlerFunc {
	return cachecontrol.New(cachecontrol.Config{
		// The must-revalidate response directive indicates that
		// the response can be stored in caches and can be reused
		// while fresh. If the response becomes stale, it must be
		// validated with the origin server before reuse.
		//
		// Typically, must-revalidate is used with max-age.
		//
		// Ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
		MustRevalidate: true,
		MaxAge:         cachecontrol.Duration(duration),
	})
}
