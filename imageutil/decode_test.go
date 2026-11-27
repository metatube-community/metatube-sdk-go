package imageutil

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"testing"

	"github.com/gen2brain/jpegli"
	"github.com/stretchr/testify/require"
)

// A JPEG with non-standard chroma subsampling (Y 2x2, Cb 1x1, Cr 2x2),
// which image/jpeg supports since Go 1.27.
const nonStandardSubsamplingJPEG = `
/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAwICQoJBwwKCgoNDQwOEh4TEhAQEiQaGxUeKyYtLComKSkv
NUQ6LzJAMykpO1E8QEZJTE1MLjlUWlNKWURLTEn/2wBDAQ0NDRIQEiMTEyNJMSkxSUlJSUlJSUlJSUlJ
SUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUn/wAARCAASACADASIAAhEBAyIB/8QA
GgAAAgMBAQAAAAAAAAAAAAAAAAUCBAYBA//EACkQAAIBAwMDAgcBAAAAAAAAAAECAwAEEQUSIRMxQRRh
BiIjUVKBseH/xAAYAQADAQEAAAAAAAAAAAAAAAAAAQIEA//EAB0RAAICAwADAAAAAAAAAAAAAAABAhED
ITEEIkH/2gAMAwEAAhEDEQA/AGGsXCarfRWF7K1ml2zTRosoJ3jAAyOM8HimGl65LqWjNbSBvU2shiea
UYDgZG4+c4xms2ksKSxajeosv0+mmB53ePcVo7Rrc2JezlV+sxk57nP3/lZnkZphjT0xrC70do11lLtn
lq3xBpq6UbeR0nYFC2xWK8MCeSBnsalZalpiXgdAgzIxBCgHBA/2rVqmmRSzborXqbs56YJ7e9TPobiQ
MERXXtvjA81am+szrx8qdtCyY57bR3nKb+i9qqzl1bQCOBRBGB0m42j8qWW0UaXjFY1U9Qdhiiis5Kb2
UgAv6SqjTWIUZM5ycd+BTBRh5scfJ4/dFFdlxCl1kPgAf//Z
`

// A JPEG with a 3x1 luma sampling factor, which image/jpeg does not support.
const unsupportedSubsamplingJPEG = `
/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwg
JC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIy
MjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAASACADATEAAhEBAxEB/8QA
HwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIh
MUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVW
V1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXG
x8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQF
BgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAV
YnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOE
hYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq
8vP09fb3+Pn6/9oADAMBAAIRAxEAPwDgftHvR9o96PtHvWH9l+R+rf2l5h9o96KKP7L8g/tLzMn7R70f
aPej7R71+j/2X5HwP9peYfaPeiij+y/IP7S8yhk+tGT60ZPrX2HKj5W7DJ9aKKOVBdn/2Q==
`

func TestJPEGDecode(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(nonStandardSubsamplingJPEG)
	require.NoError(t, err)

	_, err = jpeg.Decode(bytes.NewReader(data))
	require.NoError(t, err)
}

func TestJPEGDecodeFallback(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(unsupportedSubsamplingJPEG)
	require.NoError(t, err)

	var jpegErr jpeg.UnsupportedError
	_, err = jpeg.Decode(bytes.NewReader(data))
	require.ErrorAs(t, err, &jpegErr)

	_, err = jpegli.Decode(bytes.NewReader(data))
	require.NoError(t, err)

	// Decode falls back to jpegli.
	_, format, err := Decode(bytes.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, "jpeg", format)
}
