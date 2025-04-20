package baidu

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/translate"
)

var _ translate.Translator = (*Baidu)(nil)

const baiduTranslateAPI = "https://api.fanyi.baidu.com/api/trans/vip/translate"

type Baidu struct {
	AppID  string `json:"baidu-app-id"`
	AppKey string `json:"baidu-app-key"`
}

func (bd *Baidu) Translate(text, from, to string) (result string, err error) {
	var (
		resp *http.Response
		// salt & sign
		salt = strconv.Itoa(rand.Intn(0x7FFFFFFF))
		sign = md5sum(bd.AppID + text + salt + bd.AppKey)
	)
	if resp, err = fetch.Post(
		baiduTranslateAPI,
		fetch.WithURLEncodedBody(map[string]string{
			"q":     text,
			"from":  parseToSupportedLanguage(from),
			"to":    parseToSupportedLanguage(to),
			"appid": bd.AppID,
			"salt":  salt,
			"sign":  sign,
		}),
		fetch.WithHeader("Content-Type", "application/x-www-form-urlencoded"),
	); err != nil {
		return
	}
	defer resp.Body.Close()

	data := struct {
		From        string `json:"from"`
		To          string `json:"to"`
		TransResult []struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"trans_result"`
		ErrorCode string `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&data); err == nil {
		if len(data.TransResult) > 0 {
			s := strings.Builder{}
			for _, t := range data.TransResult {
				s.WriteString(t.Dst)
				s.WriteByte('\n')
			}
			result = strings.TrimSpace(s.String())
		} else {
			err = fmt.Errorf("%s: %s", data.ErrorCode, data.ErrorMsg)
		}
	}
	return
}

func md5sum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func parseToSupportedLanguage(lang string) string {
	if lang = strings.ToLower(lang); lang == "" || lang == "auto" /* auto detect */ {
		return "auto"
	}
	switch lang {
	case "zh", "chs", "zh-cn", "zh_cn", "zh-hans":
		return "zh"
	case "cht", "zh-tw", "zh_tw", "zh-hk", "zh_hk", "zh-hant":
		return "cht"
	case "jp", "ja":
		return "jp"
	case "kor", "ko":
		return "kor"
	case "vie", "vi":
		return "vie"
	case "spa", "es":
		return "spa"
	case "fra", "fr":
		return "fra"
	case "ara", "ar":
		return "ara"
	case "bul", "bg":
		return "bul"
	case "est", "et":
		return "est"
	case "dan", "da":
		return "dan"
	case "fin", "fi":
		return "fin"
	case "rom", "ro":
		return "rom"
	case "slo", "sl":
		return "slo"
	case "swe", "sv":
		return "swe"
	}
	return lang
}

// https://fanyi-api.baidu.com/api/trans/product/apidoc
//var baiduErrorText = map[int]string{
//	52000: "成功",
//	52001: "请求超时",
//	52002: "系统错误",
//	52003: "未授权用户",
//	54000: "必填参数为空",
//	54001: "签名错误",
//	54003: "访问频率受限",
//	54004: "账户余额不足",
//	54005: "长query请求频繁",
//	58000: "客户端IP非法",
//	58001: "译文语言方向不支持",
//	58002: "服务当前已关闭",
//	90107: "认证未通过或未生效",
//}

func init() {
	translate.Register(&Baidu{})
}
