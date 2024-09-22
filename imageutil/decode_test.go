package imageutil

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"testing"

	"github.com/gen2brain/jpegli"
	"github.com/stretchr/testify/require"
)

func TestJPEGDecode(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(base64EncodedImage)
	require.NoError(t, err)

	var jpegErr jpeg.UnsupportedError
	_, err = jpeg.Decode(bytes.NewReader(data))
	require.ErrorAs(t, err, &jpegErr)

	_, err = jpegli.Decode(bytes.NewReader(data))
	require.NoError(t, err)
}
