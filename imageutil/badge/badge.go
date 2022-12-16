package badge

import (
	"fmt"
	"image"
	"time"

	"github.com/jellydator/ttlcache/v3"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/imageutil"
)

var (
	badgeCache = ttlcache.New[string, image.Image](
		ttlcache.WithTTL[string, image.Image](30*time.Minute),
		ttlcache.WithCapacity[string, image.Image](10),
	)
	badgeFetcher = fetch.Default(nil)
)

func init() {
	// start badge cache.
	go badgeCache.Start()
}

func Badge(src image.Image, badge string) (image.Image, error) {
	var img image.Image
	if item := badgeCache.Get(badge); item != nil {
		img = item.Value()
	} else {
		resp, err := badgeFetcher.Fetch(badge)
		if err != nil {
			return nil, fmt.Errorf("fetch badge: %w", err)
		}
		defer resp.Body.Close()
		// decode badge image.
		img, _, err = image.Decode(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decode badge: %w", err)
		}
		badgeCache.Set(badge, img, ttlcache.DefaultTTL)
	}
	wmk := imageutil.Resize(img, 0, src.Bounds().Dy()/5 /* 0.2 */)
	return imageutil.Watermark(src, wmk, image.Point{}), nil
}
