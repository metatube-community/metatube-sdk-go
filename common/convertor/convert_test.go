package convertor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToCentimeters(t *testing.T) {
	for _, unit := range []struct {
		ft, in, cm int
	}{
		{5, 5, 165},
		{5, 6, 168},
		{5, 7, 170},
		{5, 8, 173},
		{5, 9, 175},
		{5, 10, 178},
	} {
		assert.Equal(t, unit.cm, ConvertToCentimeters(unit.ft, unit.in))
	}
}
