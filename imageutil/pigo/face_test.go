package pigo

import (
	"bytes"
	"embed"
	"encoding/base64"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	R "github.com/metatube-community/metatube-sdk-go/constant"
)

/*
  # Generate MD5 Hashed Images:
  setopt null_glob
  for IMG in *.jpg *.jpeg *.png; do
    gbase64 "$IMG" > "$(md5sum "$IMG" | cut -d' ' -f1)"
  done
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
		{filename: "3685c2648be7eeeaa2cef0118873a55f", position: 0.30},
		{filename: "f100611a90fa024c73132457fa77da36", position: 0.65},
		{filename: "7848e5995a58df9d063df8543c50c943", position: 0.20},
		{filename: "345a376e579ff02a518b831b1b2b4602", position: 0.20},
		{filename: "c0c99e28da91693a27de2beb6dfd7161", position: 0.50},
		{filename: "6dbe5b2d7d7056f3b60c6d05f5176529", position: 0.25},
		{filename: "369993051097480935eadf1f468eaadb", position: 0.70},
		//{filename: "d8df07a8312543f638373eb5921f896d", position: 0.10},
		//{filename: "e1c5fce943a4ba36576607eaa585b9d8", position: 0.90},
		//{filename: "e5ff5d6966391409a0fed7d3446b12aa", position: 0.60},
	} {
		t.Run(unit.filename, func(t *testing.T) {
			data, err := fs.ReadFile("assets/" + unit.filename)
			require.NoError(t, err)

			decoded, err := base64.StdEncoding.DecodeString(string(data))
			require.NoError(t, err)

			img, _, err := image.Decode(bytes.NewReader(decoded))
			require.NoError(t, err)

			pos := CalculatePosition(img, R.BackdropImageRatio, -1)
			assert.LessOrEqual(t, math.Abs(unit.position-pos), 1e-1)

			t.Logf("detected position: %f", pos)
		})
	}
}
