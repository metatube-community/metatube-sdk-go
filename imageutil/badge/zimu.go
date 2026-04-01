package badge

import (
	"bytes"
	_ "embed"

	"github.com/jellydator/ttlcache/v3"

	"github.com/metatube-community/metatube-sdk-go/imageutil"
)

//go:embed zimu.png
var zimu []byte

//go:embed u.png
var u []byte

//go:embed uc.png
var uc []byte

func init() {
	registerEmbeddedBadge("zimu.png", zimu)
	registerEmbeddedBadge("u.png", u)
	registerEmbeddedBadge("uc.png", uc)
}

func registerEmbeddedBadge(name string, raw []byte) {
	badge, _, err := imageutil.Decode(bytes.NewReader(raw))
	if err != nil || badge == nil {
		return
	}
	badgeCache.Set(name, badge, ttlcache.NoTTL)
}
