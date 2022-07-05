package number

import (
	"fmt"
	"regexp"
	"strings"
)

var shiroutoList = []string{
	"ara",
	"bnjc",
	"dcv",
	"endx",
	"eva",
	"ezd",
	"gana",
	"hamenets",
	"hmdn",
	"hoi",
	"imdk",
	"ion",
	"jac",
	"jkz",
	"jotk",
	"ksko",
	"luxu",
	"maan",
	"mium",
	"mntj",
	"nama",
	"ntk",
	"nttr",
	"obut",
	"ore",
	"orebms",
	"orec",
	"oreco",
	"orerb",
	"oretd",
	"orex",
	"per",
	"pkjd",
	"scp",
	"scute",
	"cute",
	"shyn",
	"simm",
	"siro",
	"srcn",
	"sqb",
	"sweet",
	"svmm",
	"urf",
}

var shiroutoRe = regexp.MustCompile(fmt.Sprintf(`(?i)(%s)[-_\d]`, strings.Join(shiroutoList, "|")))
