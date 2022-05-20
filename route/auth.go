package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TokenStore map[string]struct{}

func NewTokenStore(tokens ...string) TokenStore {
	store := make(TokenStore)
	store.Add(tokens...)
	return store
}

func (store TokenStore) Add(tokens ...string) {
	for _, token := range tokens {
		store[token] = struct{}{}
	}
}

func (store TokenStore) Del(token string) {
	delete(store, token)
}

func (store TokenStore) Valid(token string) (ok bool) {
	_, ok = store[token]
	return
}

func authentication(store TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(store) == 0 {
			c.Next()
			return
		}

		header := c.GetHeader("Authorization")
		bearer, token, found := strings.Cut(header, " ")

		hasInvalidHeader := bearer != "Bearer"
		hasInvalidSecret := !found || !store.Valid(token)
		if hasInvalidHeader || hasInvalidSecret {
			abortWithStatusMessage(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized))
			return
		}

		c.Next()
	}
}
