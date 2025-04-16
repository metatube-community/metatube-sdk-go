package pigo

import (
	"image"
	"math"
	"slices"
	"sort"

	pigo "github.com/esimov/pigo/core"
)

type vote struct {
	count  int
	sumPos float64
	weight float32
}

func (v vote) avgPos() float64 {
	return v.sumPos / float64(v.count)
}

func calculateWeight(face pigo.Detection) float32 {
	return float32(face.Scale) * face.Q
}

func calculateFacePosition(img image.Image, ratio float64, face pigo.Detection) (pos float64) {
	var (
		width  = img.Bounds().Dx()
		height = img.Bounds().Dy()
	)
	if int(float64(height)*ratio) < width {
		pos = float64(face.Col) / float64(width)
	} else {
		pos = float64(face.Row) / float64(height)
	}
	return
}

func aggregateVotesFromFaces(img image.Image, ratio float64, faces []pigo.Detection) []vote {
	const step = 5 /* Range: [2,20] */
	const maxBins = 100/step + 1
	votes := make([]vote, maxBins)

	for _, face := range faces {
		pos := calculateFacePosition(img, ratio, face)
		idx := int(math.Round(pos * 100 / step))

		if idx < 0 || idx >= maxBins {
			continue // skip
		}

		v := votes[idx]
		v.count++
		v.sumPos += pos
		v.weight += calculateWeight(face)
		votes[idx] = v
	}

	// filter empty votes.
	return slices.DeleteFunc(votes, func(v vote) bool {
		return v.weight == 0
	})
}

func getTopVotedPosition(votes []vote) (float64, bool) {
	if len(votes) == 0 {
		return 0, false
	}
	// sort votes by weight.
	sort.SliceStable(votes, func(i, j int) bool {
		return votes[i].weight > votes[j].weight
	})
	return votes[0].avgPos(), true
}
