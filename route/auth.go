package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/errors"
	"github.com/metatube-community/metatube-sdk-go/route/auth"
)

func authentication(v auth.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if v != nil /* auth enabled */ {
			header := c.GetHeader("Authorization")
			bearer, token, found := strings.Cut(header, " ")

			hasInvalidHeader := bearer != "Bearer"
			hasInvalidSecret := !found || !v.Valid(token)
			if hasInvalidHeader || hasInvalidSecret {
				abortWithError(c, errors.FromCode(http.StatusUnauthorized))
				return
			}
		}
		c.Next()
	}
}
