package position

import (
	"fmt"
	"math"

	"github.com/metatube-community/metatube-sdk-go/common/cluster"
)

var _ cluster.Locatable[Position, float64] = (*Position)(nil)

type Position float64

func (p Position) DistanceTo(o Position) float64 {
	return math.Abs(float64(p) - float64(o))
}

func (p Position) IsValid() bool {
	return 0.0 <= p && p <= 1.0
}

func (p Position) String() string {
	return fmt.Sprintf("%.2f", p)
}
