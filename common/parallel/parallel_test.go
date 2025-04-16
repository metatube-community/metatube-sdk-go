package parallel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParallel(t *testing.T) {
	fn := func(s string) int {
		return len(s)
	}

	for _, unit := range []struct {
		args []string
		want []int
	}{
		{
			args: []string{"a", "ab", "hello", "world"},
			want: []int{1, 2, 5, 5},
		},
		{
			args: []string{},
			want: []int{},
		},
	} {
		result := Parallel(fn, unit.args...)
		assert.Equal(t, unit.want, result)
	}
}

func mockSlowFn(x int) int {
	time.Sleep(10 * time.Millisecond)
	return x * x
}

func BenchmarkParallel(b *testing.B) {
	args := make([]int, 100)
	for i := 0; i < 100; i++ {
		args[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Parallel(mockSlowFn, args...)
	}
}

func BenchmarkSerial(b *testing.B) {
	args := make([]int, 100)
	for i := 0; i < 100; i++ {
		args[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := make([]int, len(args))
		for j, v := range args {
			results[j] = mockSlowFn(v)
		}
	}
}
