package parallel

import (
	"sync"
)

func Parallel[T any, R any](fn func(T) R, args ...T) []R {
	var wg sync.WaitGroup
	results := make([]R, len(args))

	for i, v := range args {
		wg.Add(1)
		go func(i int, v T) {
			defer wg.Done()
			results[i] = fn(v)
		}(i, v)
	}

	wg.Wait()
	return results
}
