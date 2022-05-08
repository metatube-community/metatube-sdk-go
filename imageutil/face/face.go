package face

import (
	"image"

	pigo "github.com/esimov/pigo/core"
)

var classifier *pigo.Pigo

func init() {
	classifier, _ = pigo.NewPigo().Unpack(cascade)
}

func DetectFaces(img image.Image) (dets []pigo.Detection) {
	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
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
