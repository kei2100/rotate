package state

import (
	"sync/atomic"
)

const (
	stateNotRotating = iota
	stateRotating
	stateClosed
)

// State of the rotate.Writer
type State struct {
	// openedAt Unix time
	openedAt int64
	// file size (when opened) + written bytes
	size  int64
	state uint32
}

// NewState creates a State
func NewState(openedAt int64, size int64) *State {
	return &State{openedAt: openedAt, size: size}
}

// OpenedAt returns openedAt
func (s *State) OpenedAt() int64 {
	return atomic.LoadInt64(&s.openedAt)
}

// Size returns `file size (when opened) + written bytes`
func (s *State) Size() int64 {
	return atomic.LoadInt64(&s.size)
}

// AddSize atomically
func (s *State) AddSize(value int64) {
	atomic.AddInt64(&s.size, value)
}

// CompareAndSwapAsRotating cas not-rotating to rotating
func (s *State) CompareAndSwapAsRotating() bool {
	return atomic.CompareAndSwapUint32(&s.state, stateNotRotating, stateRotating)
}

// CompareAndSwapAsNotRotating cas rotating to not-rotating
func (s *State) CompareAndSwapAsNotRotating() bool {
	return atomic.CompareAndSwapUint32(&s.state, stateRotating, stateNotRotating)
}

// StoreAsClosed set state to closed atomically
func (s *State) StoreAsClosed() {
	atomic.StoreUint32(&s.state, stateClosed)
}

// IsClosed reports whether the state is closed
func (s *State) IsClosed() bool {
	return atomic.LoadUint32(&s.state) == stateClosed
}
