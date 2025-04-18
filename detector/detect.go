package detector

import (
	"image"
	"image/color"
	"math"

	"github.com/disintegration/imaging"
	pigo "github.com/esimov/pigo/core"

	"github.com/metatube-community/metatube-sdk-go/common/cluster"
	"github.com/metatube-community/metatube-sdk-go/common/parallel"
	"github.com/metatube-community/metatube-sdk-go/detector/internal/position"
	"github.com/metatube-community/metatube-sdk-go/detector/internal/utils"
)

const (
	maxImageWidth = 650
	minFaceSize   = 20
	maxFaceSize   = maxImageWidth * 0.8
)

var classifier *pigo.Pigo

func init() {
	classifier, _ = pigo.NewPigo().Unpack(cascade)
}

func detectFaces(params *pigo.CascadeParams, angles ...float64) []pigo.Detection {
	// initialize angles if empty.
	if len(angles) == 0 {
		angles = []float64{0.0}
	}

	detect := func(angle float64) []pigo.Detection {
		// Run the classifier over the obtained leaf nodes and return the detection results.
		// The result contains quadruplets representing the row, column, scale and detection score.
		return classifier.RunCascade(*params, angle)
	}
	return parallel.Flatten(parallel.Parallel(detect, angles...))
}

func DetectFaces(img image.Image, angles ...float64) []pigo.Detection {
	imgParams := pigo.ImageParams{
		Pixels: pigo.RgbToGrayscale(img),
		Rows:   img.Bounds().Dy(),
		Cols:   img.Bounds().Dx(),
		Dim:    img.Bounds().Dx(),
	}
	for _, params := range []pigo.CascadeParams{
		{
			MinSize:     minFaceSize,
			MaxSize:     maxFaceSize,
			ShiftFactor: 0.10,
			ScaleFactor: 1.08,
			ImageParams: imgParams,
		},
		/*
			// extra params for better accuracy.
			{
				MinSize:     minFaceSize,
				MaxSize:     maxFaceSize,
				ShiftFactor: 0.09,
				ScaleFactor: 1.0,
				ImageParams: imgParams,
			},
		*/
	} {
		if faces := detectFaces(&params, angles...); len(faces) > 0 {
			return faces
		}
	}
	return nil
}

func DetectFacesWithRotation(img image.Image, rotatedAngle float64, angles ...float64) []pigo.Detection {
	var (
		origWidth  = img.Bounds().Dx()
		origHeight = img.Bounds().Dy()
	)
	rotatedImg := imaging.Rotate(img, rotatedAngle, color.Transparent)
	faces := DetectFaces(rotatedImg, angles...)
	if rotatedAngle == 0 {
		return faces
	}
	// calculate converted coordinates.
	for i := range faces {
		x, y := utils.RotatePoint(
			faces[i].Col, faces[i].Row,
			rotatedImg.Bounds().Dx(),
			rotatedImg.Bounds().Dy(),
			math.Mod(360-rotatedAngle, 360), /* inverse angle */
		)
		x = max(min(x, origWidth), 0)
		y = max(min(y, origHeight), 0)
		faces[i].Col, faces[i].Row = x, y
	}
	return faces
}

func DetectFacesWithMultiAngles(img image.Image) []pigo.Detection {
	fixedAngles := []float64{ // in radians
		0.00,
		0.13,
		0.87,
	}
	rotatedAngles := []float64{ // in degrees
		0,
		90,
		270,
	}
	detect := func(angle float64) []pigo.Detection {
		return DetectFacesWithRotation(img, angle, fixedAngles...)
	}
	return parallel.Flatten(parallel.Parallel(detect, rotatedAngles...))
}

func DetectPrimaryFacePosition(img image.Image, ratio float64, debugs ...debugFunc) (float64, bool) {
	// limit max width for performance improvement.
	if img.Bounds().Dx() > maxImageWidth {
		img = imaging.Resize(
			img, maxImageWidth, 0,
			imaging.NearestNeighbor, /* fastest */
		)
	}
	// detect faces from different angles.
	faces := DetectFacesWithMultiAngles(img)
	dim := 0
	// calculate pos-vector groups based on distances.
	groups := clusterFacesToGroups(img, faces, dim /* X */)
	// callback debug functions.
	defer func() {
		for _, fn := range debugs {
			fn(img, faces, groups)
		}
	}()
	vec, ok := topWeightedVector(groups)
	if !ok || vec.Dim() != 1 {
		return 0, false
	}
	return float64(vec.At(dim)), true
}

// debugFunc should be used for debugging only.
type debugFunc func(image.Image, []pigo.Detection, []cluster.Group[position.WeightedVector, float64])
