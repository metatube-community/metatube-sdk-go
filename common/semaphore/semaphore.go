package semaphore

// Semaphore struct represents a semaphore.
type Semaphore struct {
	ch chan struct{}
}

// New creates a new semaphore with the specified capacity.
func New(capacity int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, capacity)}
}

// Acquire blocks until it can acquire a semaphore token.
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// Release releases a semaphore token.
func (s *Semaphore) Release() {
	<-s.ch
}
