package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"

	"github.com/javtube/javtube-sdk-go/translate"
)

const defaultMaxRPS = 1

type translateEngine = string

const (
	googleTranslateEngine translateEngine = "google"
	baiduTranslateEngine  translateEngine = "baidu"
)

type translateQuery struct {
	Q      string `form:"q" binding:"required"`
	From   string `form:"from"`
	To     string `form:"to" binding:"required"`
	Engine string `form:"engine"`
}

func getTranslate(rate int) gin.HandlerFunc {
	limiter := ratelimit.New(rate)
	return func(c *gin.Context) {
		query := &translateQuery{
			From:   "auto",
			Engine: googleTranslateEngine,
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		// apply limit
		limiter.Take()

		var (
			result string
			err    error
		)
		switch strings.ToLower(query.Engine) {
		case googleTranslateEngine:
			result, err = translate.GoogleTranslate(query.Q, query.From, query.To,
				c.Query("google-api-key"))
		case baiduTranslateEngine:
			result, err = translate.BaiduTranslate(query.Q, query.From, query.To,
				c.Query("baidu-app-id"), c.Query("baidu-app-key"))
		default:
			abortWithStatusMessage(c, http.StatusBadRequest, "invalid translate engine")
			return
		}

		// TODO: response struct
		c.JSON(http.StatusOK, gin.H{
			"error":  err.Error(),
			"result": result,
		})
	}
}
