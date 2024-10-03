package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/schema"

	"github.com/metatube-community/metatube-sdk-go/translate"
	_ "github.com/metatube-community/metatube-sdk-go/translate/baidu"
	_ "github.com/metatube-community/metatube-sdk-go/translate/deepl"
	_ "github.com/metatube-community/metatube-sdk-go/translate/google"
	_ "github.com/metatube-community/metatube-sdk-go/translate/googlefree"
	_ "github.com/metatube-community/metatube-sdk-go/translate/openai"
)

type translateQuery struct {
	Q      string `form:"q" binding:"required"`
	From   string `form:"from"`
	To     string `form:"to" binding:"required"`
	Engine string `form:"engine" binding:"required"`
}

type translateResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"translated_text"`
}

func getTranslate() gin.HandlerFunc {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag("json")
	decoder.IgnoreUnknownKeys(true)

	return func(c *gin.Context) {
		query := &translateQuery{
			From: "auto",
		}
		if err := c.ShouldBindQuery(query); err != nil {
			abortWithStatusMessage(c, http.StatusBadRequest, err)
			return
		}

		decode := func(v any) error {
			return decoder.Decode(v, c.Request.URL.Query())
		}

		result, err := translate.
			New(query.Engine, decode).
			Translate(query.Q, query.From, query.To)
		if err != nil {
			abortWithError(c, err)
			return
		}

		c.JSON(http.StatusOK, &responseMessage{
			Data: &translateResponse{
				From: query.From,
				To:   query.To,
				Text: result,
			},
		})
	}
}
