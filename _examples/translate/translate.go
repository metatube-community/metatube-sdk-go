package main

import (
	"github.com/metatube-community/metatube-sdk-go/translate"
)

func main() {
	var (
		appId  = "XXX"
		appKey = "XXX"
	)
	// Translate `Hello` from auto to Japanese by Baidu.
	translate.BaiduTranslate("Hello", "auto", "ja", appId, appKey)

	var apiKey = "XXX"
	// Translate `Hello` from auto to simplified Chinese by Google.
	translate.GoogleTranslate("Hello", "auto", "zh-cn", apiKey)
}
