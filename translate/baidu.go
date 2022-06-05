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
	resp, err = fetch.Post(
		baiduTranslateAPI,
		fetch.WithURLEncodedBody(map[string]string{
			"q":     q,
			"from":  from,
			"to":    to,
			"appid": appID,
			"salt":  salt,
			"sign":  sign,
		}),
		fetch.WithHeader("Content-Type", "application/x-www-form-urlencoded"),
	)

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
			err = fmt.Errorf("error code: %s", data.ErrorCode)
		}
	}
	return
}

func md5sum(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
