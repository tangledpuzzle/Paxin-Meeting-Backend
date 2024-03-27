package network

import (
	"sync"
)

type BufferPool struct {
	name            string
	initialCapacity int
	bufferSize      int
	misses          int
	freeBuffers     chan []byte
	mutex           sync.Mutex
}

var pools []*BufferPool
var poolsMutex sync.Mutex

func NewBufferPool(name string, initialCapacity, bufferSize int) *BufferPool {
	pool := &BufferPool{
		name:            name,
		initialCapacity: initialCapacity,
		bufferSize:      bufferSize,
		misses:          0,
		freeBuffers:     make(chan []byte, initialCapacity),
	}

	for i := 0; i < initialCapacity; i++ {
		pool.freeBuffers <- make([]byte, bufferSize)
	}

	poolsMutex.Lock()
	defer poolsMutex.Unlock()
	pools = append(pools, pool)

	return pool
}

func (bp *BufferPool) GetInfo() (string, int, int, int, int, int) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()
	return bp.name, len(bp.freeBuffers), bp.initialCapacity, bp.initialCapacity * (1 + bp.misses), bp.bufferSize, bp.misses
}

func (bp *BufferPool) AcquireBuffer() []byte {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	select {
	case buffer := <-bp.freeBuffers:
		return buffer
	default:
		bp.misses++
		for i := 0; i < bp.initialCapacity; i++ {
			bp.freeBuffers <- make([]byte, bp.bufferSize)
		}
		return <-bp.freeBuffers
	}
}

func (bp *BufferPool) ReleaseBuffer(buffer []byte) {
	if buffer == nil {
		return
	}

	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.freeBuffers <- buffer
}

func (bp *BufferPool) Free() {
	poolsMutex.Lock()
	defer poolsMutex.Unlock()

	for i, pool := range pools {
		if pool == bp {
			pools = append(pools[:i], pools[i+1:]...)
			break
		}
	}
}

// Example for use
//   pool := NewBufferPool("TestPool", 10, 1024)

//   name, freeCount, initialCapacity, currentCapacity, bufferSize, misses := pool.GetInfo()
//   fmt.Printf("Name: %s\nFree Count: %d\nInitial Capacity: %d\nCurrent Capacity: %d\nBuffer Size: %d\nMisses: %d\n",
// 	  name, freeCount, initialCapacity, currentCapacity, bufferSize, misses)

//   buffer := pool.AcquireBuffer()
//   fmt.Println("Acquired buffer with length:", len(buffer))

//   pool.ReleaseBuffer(buffer)

//   pool.Free()
