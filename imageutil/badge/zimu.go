package badge

import (
	"bytes"
	_ "embed"
	"image"

	"github.com/jellydator/ttlcache/v3"
)

//go:embed zimu.png
var zimu []byte

func init() {
	badge, _, _ := image.Decode(bytes.NewReader(zimu))
	badgeCache.Set("zimu.png", badge, ttlcache.NoTTL)
}
