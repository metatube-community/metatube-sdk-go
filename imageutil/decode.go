//go:build !vips

package imageutil

import (
	"image"
)

var (
	Decode       = image.Decode
	DecodeConfig = image.DecodeConfig
)
