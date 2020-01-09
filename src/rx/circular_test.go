package rx

import (
	"testing"
)

func TestCircularLoopless(t *testing.T) {
	capacity := 10
	count := 5
	buffer := NewCircularBuffer(capacity)
	for i := 1; i <= count; i++ {
		buffer.Add(i)
	}

	if buffer.Capacity != capacity {
		t.Fatalf("Expected Length %v but got %v", capacity, buffer.Capacity)
	}
	if buffer.Length != capacity/2 {
		t.Fatalf("Expected Length %v but got %v", count, buffer.Length)
	}

	i, v := buffer.Next(-1)
	if i != 1 || v != 1 {
		t.Fatalf("Expected index value pair %v:%v but got %v:%v", 1, 1, i, v)
	}
}

func TestCircularLoop(t *testing.T) {
	capacity := 10
	count := 100
	buffer := NewCircularBuffer(capacity)

	if buffer.Length != 0 {
		t.Fatalf("Expected Length %v but got %v", 0, buffer.Length)
	}

	for x := 1; x < count; x++ {
		buffer.Add(x)
	}

	if buffer.Capacity != capacity {
		t.Fatalf("Expected Length %v but got %v", capacity, buffer.Capacity)
	}
	if buffer.Length != capacity {
		t.Fatalf("Expected Length %v but got %v", capacity, buffer.Length)
	}

	i := -1
	var v interface{}
	for j := 0; j < capacity; j++ {
		i, v = buffer.Next(i)
		if j == capacity && i != -1 {
			t.Fatalf("Expected end of data (-1) for item %v but got %v", capacity, i)
		}
		if i == -1 {
			break
		}
		if count-capacity+j != v {
			t.Fatalf("Expected index value pair %v:%v but got %v:%v", j, count-capacity+j, i, v)
		}
	}

	if buffer.Capacity != capacity {
		t.Fatalf("Expected Length %v but got %v", capacity, buffer.Capacity)
	}
}
