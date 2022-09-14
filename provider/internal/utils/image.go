package utils

import (
	"image"
	"sync"

	"github.com/javtube/javtube-sdk-go/common/fetch"
	"github.com/javtube/javtube-sdk-go/imageutil"
	"github.com/javtube/javtube-sdk-go/provider"
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

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func SimilarImage(imageUrl1, imageUrl2 string, fetcher provider.Fetcher) bool {
	var (
		wg         sync.WaitGroup
		img1, img2 image.Image
	)

	wg.Add(2)
	for i, imageUrl := range []string{
		imageUrl1,
		imageUrl2,
	} {
		// Async fetching.
		go func(i int, imageUrl string) {
			defer wg.Done()
			if img, err := getImageByURL(imageUrl, fetcher); err == nil {
				if i%2 == 0 {
					img1 = img
				} else {
					img2 = img
				}
			}
		}(i, imageUrl)
	}
	wg.Wait()

	if img1 == nil || img2 == nil {
		return false
	}

	img1 = imageutil.CropImagePosition(img1, 0.7, 0.5)
	img2 = imageutil.CropImagePosition(img2, 0.7, 0.5)
	return imageutil.Similar(img1, img2)
}
