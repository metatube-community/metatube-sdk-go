package route

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/engine"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

func redirect(app *engine.Engine) gin.HandlerFunc {
	const (
		separator = ":"
		queryKey  = "redirect"
	)
	return func(c *gin.Context) {
		getQuery := func(key string) (value string) {
			m, _ := parseQuery(c.Request.URL.RawQuery)
			return m.Get(key)
		}
		if redir := getQuery(queryKey); redir != "" {
			var (
				provider string
				id       string
			)
			if ss := strings.Split(redir, separator); len(ss) > 1 {
				provider, id = ss[0], ss[1]
			}

			var (
				info any
				err  error
			)
			if id, err = url.QueryUnescape(id); err != nil {
				abortWithError(c, err)
				return
			}

			switch {
			case app.IsActorProvider(provider):
				info, err = app.GetActorInfoByProviderID(provider, id, true)
			case app.IsMovieProvider(provider):
				info, err = app.GetMovieInfoByProviderID(provider, id, true)
			default:
				abortWithError(c, mt.ErrProviderNotFound)
				return
			}
			if err != nil {
				abortWithError(c, err)
				return
			}

			var homepage string
			switch v := info.(type) {
			case *model.ActorInfo:
				homepage = v.Homepage
			case *model.MovieInfo:
				homepage = v.Homepage
			}
			c.Redirect(http.StatusTemporaryRedirect, homepage)

			c.Abort() // abort pending middlewares
			return
		}
		c.Next()
	}
}

func parseQuery(query string) (m url.Values, err error) {
	m = make(url.Values)
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		m[key] = append(m[key], value)
	}
	return m, err
}
