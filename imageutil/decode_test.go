package imageutil

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"testing"

	"github.com/gen2brain/jpegli"
	"github.com/stretchr/testify/require"
)

const base64EncodedImage = `
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

func TestJPEGDecode(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(base64EncodedImage)
	require.NoError(t, err)

	var jpegErr jpeg.UnsupportedError
	_, err = jpeg.Decode(bytes.NewReader(data))
	require.ErrorAs(t, err, &jpegErr)

	_, err = jpegli.Decode(bytes.NewReader(data))
	require.NoError(t, err)
}
