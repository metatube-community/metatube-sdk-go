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

func SimilarImage(imageUrlA, imageUrlB string, fetcher provider.Fetcher) bool {
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

	imgA = imageutil.CropImagePosition(imgA, 0.7, 0.5)
	imgB = imageutil.CropImagePosition(imgB, 0.7, 0.5)
	return imageutil.Similar(imgA, imgB)
}
