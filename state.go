package rotate

import (
	"sync/atomic"
)

type FileState struct {
	openedAt int64
	// file size (when opened) + written bytes
	size int64
}

func (s *FileState) OpenedAt() int64 {
	return atomic.LoadInt64(&s.openedAt)
}

func (s *FileState) Size() int64 {
	return atomic.LoadInt64(&s.size)
}

func (s *FileState) addSize(value int64) {
	atomic.AddInt64(&s.size, value)
}

const (
	stateNotRotating = iota
	stateRotating
	stateClosed
)

type state struct {
	FileState
	state uint32
}

func (s *state) compareAndSwapAsRotating() bool {
	return atomic.CompareAndSwapUint32(&s.state, stateNotRotating, stateRotating)
}

func (s *state) compareAndSwapAsNotRotating() bool {
	return atomic.CompareAndSwapUint32(&s.state, stateRotating, stateNotRotating)
}

func (s *state) storeAsClosed() {
	atomic.StoreUint32(&s.state, stateClosed)
}

func (s *state) isClosed() bool {
	return atomic.LoadUint32(&s.state) == stateClosed
}
