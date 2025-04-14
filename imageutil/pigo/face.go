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

func detectFaces(img image.Image, params *pigo.CascadeParams) (dets []pigo.Detection) {
	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets = classifier.RunCascade(*params, 0.0)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0.2)
	return
}

func DetectFaces(img image.Image) []pigo.Detection {
	imgParams := pigo.ImageParams{
		Pixels: pigo.RgbToGrayscale(img),
		Rows:   img.Bounds().Dy(),
		Cols:   img.Bounds().Dx(),
		Dim:    img.Bounds().Dx(),
	}
	for _, params := range []pigo.CascadeParams{
		{
			MinSize:     20,
			MaxSize:     750,
			ShiftFactor: 0.1,
			ScaleFactor: 1.0,
			ImageParams: imgParams,
		},
		{
			MinSize:     20,
			MaxSize:     800,
			ShiftFactor: 0.09,
			ScaleFactor: 1.0,
			ImageParams: imgParams,
		},
	} {
		if dets := detectFaces(img, &params); len(dets) > 0 {
			return dets
		}
	}
	return nil
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
