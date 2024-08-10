package route

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/metatube-community/metatube-sdk-go/translate"
)

const (
	googleTranslateEngine     = "google"
	googleFreeTranslateEngine = "googlefree"
	baiduTranslateEngine      = "baidu"
	deeplTranslateEngine      = "deepl"
	openaiTranslateEngine     = "openai"
)

const (
	// Google
	googleAPIKey = "google-api-key"

	// DeepL
	deeplAPIKey = "deepl-api-key"

	// Openai
	openaiAPIKey = "openai-api-key"

	// Baidu
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
		case googleFreeTranslateEngine:
			result, err = translate.GoogleFreeTranslate(query.Q, query.From, query.To)
		case baiduTranslateEngine:
			result, err = translate.BaiduTranslate(query.Q, query.From, query.To,
				c.Query(baiduAPPID), c.Query(baiduAPPKey))
		case deeplTranslateEngine:
			result, err = translate.DeepLTranslate(query.Q, query.From, query.To,
				c.Query(deeplAPIKey))
		case openaiTranslateEngine:
			result, err = translate.OpenaiTranslate(query.Q, query.From, query.To,
				c.Query(openaiAPIKey))
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
