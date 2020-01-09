package rx

// This is not thread safe *but* used in a single goroutine
// and avoiding async access so it shouldn't need to worry
// about concurrency

// CircularBuffer type
type CircularBuffer struct {
	// lock     sync.RWMutex
	buffer   []interface{}
	start    int
	end      int
	Capacity int
	Length   int
}

// NewCircularBuffer init
func NewCircularBuffer(capacity int) *CircularBuffer {
	id := &CircularBuffer{
		buffer:   make([]interface{}, capacity),
		start:    0,
		end:      -1,
		Capacity: capacity,
		Length:   0,
	}
	return id
}

// Add a value to the buffer (looks good)
func (id *CircularBuffer) Add(value interface{}) {
	// defer id.lock.Unlock()
	// id.lock.Lock()

	id.end++
	// loop check
	if id.end >= id.Capacity {
		id.end = 0
	}

	// when buffer is empty (Length == 0, to be 1), allow start == end
	if id.end == id.start && id.Length != 0 {
		id.start++
	}

	// start loop check
	if id.start >= id.Capacity {
		id.start = 0
	}

	// append the value and increase Length
	id.buffer[id.end] = value
	if id.Length < id.Capacity {
		id.Length++
	}
}

// Next iterate the values
func (id *CircularBuffer) Next(next int) (int, interface{}) {
	// defer id.lock.RUnlock()
	// id.lock.RLock()

	index := next

	// -1 == new iteration
	if index == -1 {
		index = id.start
	}

	// base case, are we at buffer length 1 -or- is index out of range
	if (index == id.start && next != -1) || (index < id.start && index > id.end) {
		return -1, nil
	}

	// hold the current value
	value := id.buffer[index]

	// prepare the next index
	index++
	// loop check
	if index >= id.Capacity {
		index = 0
	}

	// return next index and current value
	return index, value
}
