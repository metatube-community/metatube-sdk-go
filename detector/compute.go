package detector

import (
	"image"

	pigo "github.com/esimov/pigo/core"

	"github.com/metatube-community/metatube-sdk-go/common/cluster"
	"github.com/metatube-community/metatube-sdk-go/detector/internal/position"
)

const (
	// tolerance defines the max distance between
	// two vectors to be considered as the same.
	tolerance = 0.05
)

func computeFaceWeight(face pigo.Detection) float64 {
	// face size (scale) * quality factor.
	return float64(face.Scale) * float64(face.Q)
}

func extractFaceVector(img image.Image, face pigo.Detection) position.Vector {
	var (
		width  = img.Bounds().Dx()
		height = img.Bounds().Dy()
	)
	return position.NewVector(
		position.Position(float64(face.Col)/float64(width)),  // X
		position.Position(float64(face.Row)/float64(height)), // Y
	)
}

func dominantAxisByRatio(img image.Image, ratio float64) int {
	var (
		width  = img.Bounds().Dx()
		height = img.Bounds().Dy()
	)
	if int(float64(height)*ratio) < width {
		return 0 // X
	} else {
		return 1 // Y
	}
}

func clusterFaceVectors(img image.Image, faces []pigo.Detection, dims ...int) []cluster.Group[position.WeightedVector, float64] {
	vecs := make([]position.WeightedVector, len(faces))
	for i, face := range faces {
		vec := extractFaceVector(img, face)
		vecs[i] = position.NewWeightedVector(
			// select vector dimensions:
			// X:(0), Y:(1), XY:(0,1).
			vec.Select(dims...),
			computeFaceWeight(face),
		)
	}
	return cluster.GroupByDistance(vecs, tolerance)
}

func getDominantVector(groups []cluster.Group[position.WeightedVector, float64]) (position.Vector, bool) {
	if len(groups) == 0 {
		return position.Vector{}, false
	}
	// sort weighted vector groups.
	cluster.SortGroupsByWeight(groups)
	// calculate the average vector.
	vec := position.WeightedAverageVector(groups[0].Items)
	return vec, true
}
