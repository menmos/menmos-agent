package ring_test

import (
	"encoding/json"
	"testing"

	"github.com/menmos/menmos-agent/agent/xecute/ring"
)

func TestBuffer_BasicReadWrite(t *testing.T) {
	buf := ring.New[int](5)

	buf.Write(1)

	if val := buf.Read(); val != 1 {
		t.Fatalf("expected read to return 1 got %d", val)
	}
}

func TestBuffer_MovingWriteHead(t *testing.T) {
	buf := ring.New[int](5)
	buf.Write(1)

	if val := buf.Read(); val != 1 {
		t.Fatalf("expected read to return 1 got %d", val)
	}

	// This moves the write head far from the read head.
	buf.Write(2)
	buf.Write(3)
	buf.Write(4)

	// Our read head shouldn't have moved
	for expected := 2; expected < 5; expected++ {
		if val := buf.Read(); val != expected {
			t.Fatalf("expected read to return %d got %d", expected, val)
		}
	}
}

func TestBuffer_OverWrite(t *testing.T) {
	buf := ring.New[int](5)

	for i := 0; i < 10; i++ {
		buf.Write(i)
	}

	for expected := 5; expected < 10; expected++ {
		if val := buf.Read(); val != expected {
			t.Fatalf("expected read to return %d got %d", expected, val)
		}
	}
}

func TestBuffer_ZeroSized(t *testing.T) {
	buf := ring.New[int](0)
	buf.Write(42)
}

func FuzzBuffer(f *testing.F) {
	raw, _ := json.Marshal(struct {
		Chose string
	}{Chose: "hello"})

	f.Add(raw)

	buf := ring.New[[]byte](5)

	f.Fuzz(func(t *testing.T, orig []byte) {
		buf.Write(orig)

		value := buf.Read()

		if len(value) != len(orig) {
			t.Fatalf("slices are of uneven lengths")
		}

		for i := 0; i < len(value); i++ {
			if value[i] != orig[i] {
				t.Fatalf("slice value at index %d differs", i)
			}
		}
	})
}
