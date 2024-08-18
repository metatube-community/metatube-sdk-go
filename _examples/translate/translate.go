package main

import (
	"github.com/metatube-community/metatube-sdk-go/translate/baidu"
	"github.com/metatube-community/metatube-sdk-go/translate/google"
)

func main() {
	var (
		appId  = "XXX"
		appKey = "XXX"
	)
	// Translate `Hello` from auto to Japanese by Baidu.
	baidu.Translate("Hello", "auto", "ja", appId, appKey)

	apiKey := "XXX"
	// Translate `Hello` from auto to simplified Chinese by Google.
	google.Translate("Hello", "auto", "zh-cn", apiKey)
}
