package pigo

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"sort"
	"testing"

	pigo "github.com/esimov/pigo/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	R "github.com/metatube-community/metatube-sdk-go/constant"
)

/*
  # Generate MD5-hashed images:
  setopt null_glob
  for IMG in *.jpg *.jpeg *.png; do
    gbase64 "$IMG" > "$(md5sum "$IMG" | cut -d' ' -f1)"
  done

  # Decode image:
  gbase64 -d "$IMG"  > "$IMG".jpg
*/

//go:embed assets/*
var fs embed.FS

func TestCalculatePosition(t *testing.T) {
	for _, unit := range []struct {
		filename string
		position float64
	}{
		{filename: "809ee47a17a7938ebd6d908244b962c8", position: 0.40},
		{filename: "c7806a2f581012eb71dce597f682c7a2", position: 0.22},
		{filename: "aa237b3a2bfd35dbe10c386c7ac777ae", position: 0.30},
		{filename: "c2a6bf02b748bb460a0dbb550f39c635", position: 0.75},
		{filename: "1d8aaf63245426c4a32720bdbf33a651", position: 0.80},
		{filename: "e953fce3bf5ec5746ead8954bec758e0", position: 0.25},
		{filename: "d054f170d52c83a773571675d954e3bb", position: 0.78},
		{filename: "3685c2648be7eeeaa2cef0118873a55f", position: 0.60},
		{filename: "7848e5995a58df9d063df8543c50c943", position: 0.20},
		{filename: "345a376e579ff02a518b831b1b2b4602", position: 0.20},
		{filename: "c0c99e28da91693a27de2beb6dfd7161", position: 0.50},
		{filename: "6dbe5b2d7d7056f3b60c6d05f5176529", position: 0.25},
		{filename: "369993051097480935eadf1f468eaadb", position: 0.70},
		{filename: "d8df07a8312543f638373eb5921f896d", position: 0.10},
		{filename: "e8e85575a04d75d2bc29abb4bb7fb447", position: 0.25},
		// Failed detection:
		// {filename: "f100611a90fa024c73132457fa77da36", position: 0.65},
		// {filename: "e1c5fce943a4ba36576607eaa585b9d8", position: 0.90},
		// {filename: "e5ff5d6966391409a0fed7d3446b12aa", position: 0.60},
	} {
		t.Run(unit.filename, func(t *testing.T) {
			data, err := fs.ReadFile("assets/" + unit.filename)
			require.NoError(t, err)

			decoded, err := base64.StdEncoding.DecodeString(string(data))
			require.NoError(t, err)

			img, _, err := image.Decode(bytes.NewReader(decoded))
			require.NoError(t, err)

			dets := DetectFacesAdvanced(img)
			printDets(t, img, R.BackdropImageRatio, dets)

			pos := CalculatePosition(img, R.BackdropImageRatio, -1, dets)
			if !assert.LessOrEqualf(t,
				math.Abs(unit.position-pos), 1e-1,
				"expect pos=%.2f, but got pos=%.2f", unit.position, pos) {
				// Debug image.
				debugImg := drawBoxes(img, dets)
				saveImage(unit.filename, debugImg)
			}

			t.Logf("detected position: %f", pos)
		})
	}
}

func printDets(t *testing.T, img image.Image, ratio float64, dets []pigo.Detection) {
	sort.SliceStable(dets, func(i, j int) bool {
		return float32(dets[i].Scale)*dets[i].Q > float32(dets[j].Scale)*dets[j].Q
	})
	for _, det := range dets {
		var (
			p float64
			w = img.Bounds().Dx()
			h = img.Bounds().Dy()
		)
		if int(float64(h)*ratio) < w {
			p = float64(det.Col) / float64(w)
		} else {
			p = float64(det.Row) / float64(h)
		}
		t.Logf("%v, weight=%.2f, pos=%.2f", det, float32(det.Scale)*det.Q, p)
	}
}

func saveImage(name string, img image.Image) {
	outputFile := fmt.Sprintf("%s.jpg", name)
	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	jpeg.Encode(f, img, nil)
}

func drawBoxes(img image.Image, dets []pigo.Detection) image.Image {
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	_ = blue
	_ = yellow

	for _, m := range dets {
		x0 := m.Col - m.Scale/2
		y0 := m.Row - m.Scale/2
		x1 := m.Col + m.Scale/2
		y1 := m.Row + m.Scale/2

		if x0 < 0 {
			x0 = 0
		}
		if y0 < 0 {
			y0 = 0
		}
		if x1 >= rgba.Bounds().Dx() {
			x1 = rgba.Bounds().Dx() - 1
		}
		if y1 >= rgba.Bounds().Dy() {
			y1 = rgba.Bounds().Dy() - 1
		}

		// Draw red rectangle
		for x := x0; x <= x1; x++ {
			rgba.Set(x, y0, red)
			rgba.Set(x, y1, red)
		}
		for y := y0; y <= y1; y++ {
			rgba.Set(x0, y, red)
			rgba.Set(x1, y, red)
		}

		// Draw label inside the box
		label := fmt.Sprintf("(%d,%d,%d,%.2f)", m.Col, m.Row, m.Scale, m.Q)
		point := fixed.Point26_6{
			X: fixed.I(x0 + 2),
			Y: fixed.I(y0 + 12),
		}
		d := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(yellow),
			Face: basicfont.Face7x13,
			Dot:  point,
		}
		d.DrawString(label)
	}

	return rgba
}
