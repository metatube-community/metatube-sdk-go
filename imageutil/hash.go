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

func AverageHashDistance(imgA, imgB image.Image) (distance int) {
	hashA, _ := goimagehash.AverageHash(imgA)
	hashB, _ := goimagehash.AverageHash(imgB)
	distance, _ = hashA.Distance(hashB)
	return
}

func DifferenceHashDistance(imgA, imgB image.Image) (distance int) {
	hashA, _ := goimagehash.DifferenceHash(imgA)
	hashB, _ := goimagehash.DifferenceHash(imgB)
	distance, _ = hashA.Distance(hashB)
	return
}

func PerceptionHashDistance(imgA, imgB image.Image) (distance int) {
	hashA, _ := goimagehash.PerceptionHash(imgA)
	hashB, _ := goimagehash.PerceptionHash(imgB)
	distance, _ = hashA.Distance(hashB)
	return
}

func Similar(imgA, imgB image.Image) bool {
	switch {
	case AverageHashDistance(imgA, imgB) < thAverageHash:
		return true
	case DifferenceHashDistance(imgA, imgB) < thDifferenceHash:
		return true
	case PerceptionHashDistance(imgA, imgB) < thPerceptionHash:
		return true
	default:
		return false
	}
}
