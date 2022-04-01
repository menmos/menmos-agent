package ring

import (
	"sync"
)

type Buffer[T any] struct {
	entries   []T
	size      uint32
	writeHead uint32
	readHead  uint32
	mutex     *sync.Mutex
}

func New[T any](size uint32) *Buffer[T] {
	return &Buffer[T]{
		entries:   make([]T, size),
		size:      size,
		writeHead: 0,
		readHead:  0,
		mutex:     &sync.Mutex{},
	}

}

func (b *Buffer[T]) Write(value T) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.size == 0 {
		return
	}

	b.entries[b.writeHead] = value

	b.writeHead += 1
	if b.writeHead >= b.size {
		b.writeHead = 0
	}
}

func (b *Buffer[T]) Read() T {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.size == 0 {
		// Return the 0 value
		var val T
		return val
	}

	val := b.entries[b.readHead]

	b.readHead += 1
	if b.readHead >= b.size {
		b.readHead = 0
	}

	return val
}
