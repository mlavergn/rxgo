package rx

import (
	"testing"
)

func TestCircularLoopless(t *testing.T) {
	circ := NewCircularBuffer(4)
	circ.Add(1)
	circ.Add(2)

	if circ.Capacity != 4 {
		t.Fatalf("Expected Length %v but got %v", 4, circ.Capacity)
	}
	if circ.Length != 2 {
		t.Fatalf("Expected Length %v but got %v", 2, circ.Length)
	}

	i, v := circ.Next(-1)
	if i != 1 || v != 1 {
		t.Fatalf("Expected index value pair %v:%v but got %v:%v", 1, 1, i, v)
	}
}

func TestCircularLoop(t *testing.T) {
	size := 4
	circ := NewCircularBuffer(size)
	circ.Add(1)
	circ.Add(2)
	circ.Add(3)
	circ.Add(4)
	circ.Add(5)
	circ.Add(6)

	if circ.Capacity != size {
		t.Fatalf("Expected Length %v but got %v", size, circ.Capacity)
	}
	if circ.Length != size {
		t.Fatalf("Expected Length %v but got %v", size, circ.Length)
	}

	i := -1
	var v interface{}
	for j := 0; j < size; j++ {
		i, v = circ.Next(i)
		if j == size && i != -1 {
			t.Fatalf("Expected end of data (-1) for item %v but got %v", size, i)
		}
		if i == -1 {
			break
		}
		if j+3 != v {
			t.Fatalf("Expected index value pair %v:%v but got %v:%v", j, j+3, i, v)
		}
	}

	if circ.Capacity != size {
		t.Fatalf("Expected Length %v but got %v", size, circ.Capacity)
	}
}
