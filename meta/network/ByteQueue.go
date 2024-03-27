package network

type ByteQueue struct {
	head   int
	tail   int
	size   int
	buffer []byte
}

func NewByteQueue() *ByteQueue {
	return &ByteQueue{
		buffer: make([]byte, 2048),
	}
}

func (bq *ByteQueue) Clear() {
	bq.head = 0
	bq.tail = 0
	bq.size = 0
}

func (bq *ByteQueue) GetPacketID() byte {
	if bq.size >= 1 {
		return bq.buffer[bq.head]
	}
	return 0xFF
}

func (bq *ByteQueue) GetPacketLength() int {
	if bq.size >= 3 {
		return int(bq.buffer[(bq.head+1)%len(bq.buffer)])<<8 | int(bq.buffer[(bq.head+2)%len(bq.buffer)])
	}
	return 0
}

func (bq *ByteQueue) Dequeue(buffer []byte, offset, size int) (int, error) {
	if size > bq.size {
		size = bq.size
	}

	if size == 0 {
		return 0, nil
	}

	if bq.head < bq.tail {
		copy(buffer[offset:offset+size], bq.buffer[bq.head:bq.head+size])
	} else {
		rightLength := len(bq.buffer) - bq.head

		if rightLength >= size {
			copy(buffer[offset:offset+size], bq.buffer[bq.head:bq.head+size])
		} else {
			copy(buffer[offset:offset+rightLength], bq.buffer[bq.head:])
			copy(buffer[offset+rightLength:], bq.buffer[:size-rightLength])
		}
	}

	bq.head = (bq.head + size) % len(bq.buffer)
	bq.size -= size

	if bq.size == 0 {
		bq.head = 0
		bq.tail = 0
	}

	return size, nil
}

func (bq *ByteQueue) Enqueue(buffer []byte, offset, size int) error {
	if (bq.size + size) > len(bq.buffer) {
		err := bq.setCapacity((bq.size + size + 2047) &^ 2047)
		if err != nil {
			return err
		}
	}

	if bq.head < bq.tail {
		rightLength := len(bq.buffer) - bq.tail

		if rightLength >= size {
			copy(bq.buffer[bq.tail:], buffer[offset:offset+size])
		} else {
			copy(bq.buffer[bq.tail:], buffer[offset:offset+rightLength])
			copy(bq.buffer[:size-rightLength], buffer[offset+rightLength:])
		}
	} else {
		copy(bq.buffer[bq.tail:], buffer[offset:offset+size])
	}

	bq.tail = (bq.tail + size) % len(bq.buffer)
	bq.size += size

	return nil
}

func (bq *ByteQueue) setCapacity(capacity int) error {
	newBuffer := make([]byte, capacity)

	if bq.size > 0 {
		if bq.head < bq.tail {
			copy(newBuffer, bq.buffer[bq.head:bq.head+bq.size])
		} else {
			rightLength := len(bq.buffer) - bq.head
			copy(newBuffer, bq.buffer[bq.head:])
			copy(newBuffer[rightLength:], bq.buffer[:bq.tail])
		}
	}

	bq.head = 0
	bq.tail = bq.size
	bq.buffer = newBuffer

	return nil
}

// Size возвращает размер очереди
func (bq *ByteQueue) Size() int {
	return bq.size
}
