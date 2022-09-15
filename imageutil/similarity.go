package imageutil

import (
	"image"

	"github.com/corona10/goimagehash"
)

const (
	thAverageHash    = 6
	thDifferenceHash = 5
	thPerceptionHash = 6
)

// Similar is a shortcut of images4.Similar.
func Similar(imgA, imgB image.Image) bool {
	switch {
	case averageHashDistance(imgA, imgB) < thAverageHash:
		return true
	case differenceHashDistance(imgA, imgB) < thDifferenceHash:
		return true
	case perceptionHashDistance(imgA, imgB) < thPerceptionHash:
		return true
	default:
		return false
	}
}

func averageHashDistance(imgA, imgB image.Image) (distance int) {
	hashA, _ := goimagehash.AverageHash(imgA)
	hashB, _ := goimagehash.AverageHash(imgB)
	distance, _ = hashA.Distance(hashB)
	return
}

func differenceHashDistance(imgA, imgB image.Image) (distance int) {
	hashA, _ := goimagehash.DifferenceHash(imgA)
	hashB, _ := goimagehash.DifferenceHash(imgB)
	distance, _ = hashA.Distance(hashB)
	return
}

func perceptionHashDistance(imgA, imgB image.Image) (distance int) {
	hashA, _ := goimagehash.PerceptionHash(imgA)
	hashB, _ := goimagehash.PerceptionHash(imgB)
	distance, _ = hashA.Distance(hashB)
	return
}
