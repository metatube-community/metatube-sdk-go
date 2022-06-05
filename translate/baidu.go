package translate

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/javtube/javtube-sdk-go/common/fetch"
)

const baiduTranslateAPI = "https://api.fanyi.baidu.com/api/trans/vip/translate"

func BaiduTranslate(q, from, to, appID, appKey string) (result string, err error) {
	var (
		resp *http.Response
		// salt & sign
		salt = strconv.Itoa(rand.Intn(0x7FFFFFFF))
		sign = md5sum(appID + q + salt + appKey)
	)
	if resp, err = fetch.Post(
		baiduTranslateAPI,
		fetch.WithURLEncodedBody(map[string]string{
			"q":     q,
			"from":  parseToBaiduSupportedLanguage(from),
			"to":    parseToBaiduSupportedLanguage(to),
			"appid": appID,
			"salt":  salt,
			"sign":  sign,
		}),
		fetch.WithHeader("Content-Type", "application/x-www-form-urlencoded"),
	); err != nil {
		return
	}

	data := struct {
		From        string `json:"from"`
		To          string `json:"to"`
		ErrorCode   string `json:"error_code"`
		TransResult []struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"trans_result"`
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
			err = fmt.Errorf("baidu fanyi error code: %s", data.ErrorCode)
		}
	}
	return
}

func md5sum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func parseToBaiduSupportedLanguage(lang string) string {
	if lang = strings.ToLower(lang); lang == "" || lang == "auto" /* auto detect */ {
		return "auto"
	}
	switch lang {
	case "zh", "zh-cn", "zh_cn", "zh-hans":
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

func init() {
	rand.Seed(time.Now().UnixNano())
}
