package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/route/validator"
)

func authentication(v validator.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if v != nil /* auth enabled */ {
			header := c.GetHeader("Authorization")
			bearer, token, found := strings.Cut(header, " ")

			hasInvalidHeader := bearer != "Bearer"
			hasInvalidSecret := !found || !v.Valid(token)
			if hasInvalidHeader || hasInvalidSecret {
				abortWithStatusMessage(c, http.StatusUnauthorized,
					http.StatusText(http.StatusUnauthorized))
				return
			}
		}
		c.Next()
	}
}
