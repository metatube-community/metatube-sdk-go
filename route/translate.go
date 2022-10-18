package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/javtube/javtube-sdk-go/translate"
)

const (
	googleTranslateEngine = "google"
	baiduTranslateEngine  = "baidu"
	deeplTranslateEngine  = "deepl"
)

const (
	// GoogleAPI extra query
	googleAPIKey = "google-api-key"

	deeplAPIKey = "deepl-api-key"

	// BaiduAPI extra query
	baiduAPPID  = "baidu-app-id"
	baiduAPPKey = "baidu-app-key"
)

type translateQuery struct {
	Q      string `form:"q" binding:"required"`
	From   string `form:"from"`
	To     string `form:"to" binding:"required"`
	Engine string `form:"engine" binding:"required"`
}

func getTranslate() gin.HandlerFunc {
	return func(c *gin.Context) {
		query := &translateQuery{
			From: "auto",
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		var (
			result string
			err    error
		)
		switch strings.ToLower(query.Engine) {
		case googleTranslateEngine:
			result, err = translate.GoogleTranslate(query.Q, query.From, query.To,
				c.Query(googleAPIKey))
		case baiduTranslateEngine:
			result, err = translate.BaiduTranslate(query.Q, query.From, query.To,
				c.Query(baiduAPPID), c.Query(baiduAPPKey))
		case deeplTranslateEngine:
			result, err = translate.DeeplTranslate(query.Q, query.From, query.To,
				c.Query(deeplAPIKey))
		default:
			abortWithStatusMessage(c, http.StatusBadRequest, "invalid translate engine")
			return
		}
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.JSON(http.StatusOK, &responseMessage{
			Data: gin.H{
				"from":            query.From,
				"to":              query.To,
				"translated_text": result,
			},
		})
	}
}
