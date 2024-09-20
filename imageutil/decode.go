package imageutil

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/jpeg"
	"io"

	"github.com/gen2brain/jpegli"
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

// Ref: https://github.com/gen2brain/jpegli/issues/4
func init() {
	defer func() { recover() }()
	data, _ := base64.StdEncoding.DecodeString(base64EncodedImage)
	_, _ = jpegli.Decode(bytes.NewBuffer(data))
}

func Decode(r io.Reader) (image.Image, string, error) {
	var (
		buf     bytes.Buffer
		jpegErr jpeg.UnsupportedError
	)
	m, f, err := image.Decode(io.TeeReader(r, &buf))
	if err != nil && errors.As(err, &jpegErr) {
		// Fallback to decode with jpegli.
		m, err = jpegli.Decode(io.MultiReader(&buf, r))
	}
	return m, f, err
}
