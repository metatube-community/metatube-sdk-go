package random

import (
	"strings"

	"github.com/projectdiscovery/useragent"
	sliceutil "github.com/projectdiscovery/utils/slice"
)

var _userAgents []*useragent.UserAgent

func init() {
	for _, ua := range useragent.UserAgents {
		if filter(ua) {
			_userAgents = append(_userAgents, ua)
		}
	}
}

func filter(ua *useragent.UserAgent) bool {
	return useragent.Computer(ua) &&
		!useragent.Mobile(ua) &&
		!useragent.Bot(ua) &&
		!useragent.GoogleBot(ua) &&
		!strings.Contains(ua.Raw, "Mobile")
}

func UserAgent() string {
	return sliceutil.PickRandom(_userAgents).String()
}
