package pigo

import (
	"image"
	"sort"

	pigo "github.com/esimov/pigo/core"
)

var classifier *pigo.Pigo

func init() {
	classifier, _ = pigo.NewPigo().Unpack(cascade)
}

func DetectFaces(img image.Image) (dets []pigo.Detection) {
	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     2000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: pigo.RgbToGrayscale(img),
			Rows:   img.Bounds().Dy(),
			Cols:   img.Bounds().Dx(),
			Dim:    img.Bounds().Dx(),
		},
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets = classifier.RunCascade(cParams, 0.0)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0.2)
	return
}

func CalculatePosition(img image.Image, ratio float64, pos float64) float64 {
	if dets := DetectFaces(img); len(dets) > 0 {
		sort.SliceStable(dets, func(i, j int) bool {
			return float32(dets[i].Scale)*dets[i].Q > float32(dets[j].Scale)*dets[j].Q
		})
		var (
			width  = img.Bounds().Dx()
			height = img.Bounds().Dy()
		)
		if int(float64(height)*ratio) < width {
			pos = float64(dets[0].Col) / float64(width)
		} else {
			pos = float64(dets[0].Row) / float64(height)
		}
	}
	return pos
}
