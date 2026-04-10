package pipeline

import "sync"

// RingBuffer is a thread-safe circular buffer that stores the last N lines.
type RingBuffer struct {
	mu    sync.Mutex
	lines []string
	size  int
	pos   int
	count int
}

// NewRingBuffer creates a ring buffer that holds at most n lines.
func NewRingBuffer(n int) *RingBuffer {
	return &RingBuffer{
		lines: make([]string, n),
		size:  n,
	}
}

// Write adds a line to the buffer, overwriting the oldest if full.
func (rb *RingBuffer) Write(line string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.lines[rb.pos] = line
	rb.pos = (rb.pos + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	}
}

// Lines returns the buffered lines in chronological order.
func (rb *RingBuffer) Lines() []string {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.count == 0 {
		return nil
	}
	result := make([]string, rb.count)
	start := rb.pos - rb.count
	if start < 0 {
		start += rb.size
	}
	for i := 0; i < rb.count; i++ {
		result[i] = rb.lines[(start+i)%rb.size]
	}
	return result
}
