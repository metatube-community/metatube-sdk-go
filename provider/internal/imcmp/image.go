package imcmp

import (
	"image"
	"math"
	"sync"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
	"github.com/metatube-community/metatube-sdk-go/imageutil"
	"github.com/metatube-community/metatube-sdk-go/provider"
)

var imageFetcher = fetch.Default(nil)

func getImageByURL(url string, fetcher provider.Fetcher) (image.Image, error) {
	if fetcher == nil /* default */ {
		fetcher = imageFetcher
	}

	resp, err := fetcher.Fetch(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := imageutil.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func Similar(imageUrlA, imageUrlB string, fetcher provider.Fetcher) bool {
	var (
		wg         sync.WaitGroup
		imgA, imgB image.Image
	)

	wg.Add(2)
	for i, imageUrl := range []string{
		imageUrlA,
		imageUrlB,
	} {
		// Async fetching.
		go func(i int, imageUrl string) {
			defer wg.Done()
			if img, err := getImageByURL(imageUrl, fetcher); err == nil {
				if i%2 == 0 {
					imgA = img
				} else {
					imgB = img
				}
			}
		}(i, imageUrl)
	}
	wg.Wait()

	if imgA == nil || imgB == nil {
		return false
	}

	ratioA := float64(imgA.Bounds().Dx()) / float64(imgA.Bounds().Dy())
	ratioB := float64(imgB.Bounds().Dx()) / float64(imgB.Bounds().Dy())

	const (
		tol = 0.1 // tolerance
		pos = 0.5 // center
	)

	if math.Abs(ratioA-ratioB) > tol {
		return false
	}

	if ratioA < ratioB {
		imgB = imageutil.CropImagePosition(imgB, ratioA, pos)
	} else {
		imgA = imageutil.CropImagePosition(imgA, ratioB, pos)
	}

	return imageutil.Similar(imgA, imgB)
}
