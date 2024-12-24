package route

import (
	"time"

	"github.com/gin-gonic/gin"
	cachecontrol "go.eigsys.de/gin-cachecontrol/v2"
)

func cachePublicSMaxAge(duration time.Duration) gin.HandlerFunc {
	return cachecontrol.New(cachecontrol.Config{
		Public:  true,
		SMaxAge: cachecontrol.Duration(duration),
	})
}

func cacheNoStore() gin.HandlerFunc {
	return cachecontrol.New(cachecontrol.Config{
		// The no-store response directive indicates that any
		// caches of any kind (private or shared) should not
		// store this response.
		NoStore: true,
	})
}
