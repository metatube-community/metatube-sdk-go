package vips

import (
	"bytes"
	"image/jpeg"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/metatube-community/metatube-sdk-go/common/fetch"
)

func TestImageDecode(t *testing.T) {
	resp, err := fetch.Get("https://www.1pondo.tv/moviepages/090124_001/images/str.jpg")
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var jpegErr jpeg.UnsupportedError
	_, err = jpeg.Decode(bytes.NewBuffer(data))
	require.ErrorAs(t, err, &jpegErr)

	img, err := Decode(bytes.NewBuffer(data))
	require.NoError(t, err)
	require.NotNil(t, img)

	cfg, err := DecodeConfig(bytes.NewBuffer(data))
	require.NoError(t, err)
	require.NotEmpty(t, cfg)
}
