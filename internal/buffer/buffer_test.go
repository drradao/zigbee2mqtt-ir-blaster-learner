package buffer

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	b := New()
	if b.Len() != 0 {
		t.Fatalf("expected Len=0, got %d", b.Len())
	}
	all := b.All()
	if all == nil {
		t.Fatal("All() returned nil, want non-nil empty slice")
	}
	if len(all) != 0 {
		t.Fatalf("expected empty slice, got len=%d", len(all))
	}
}

func TestAdd_Single(t *testing.T) {
	b := New()
	msg := IRMessage{Payload: []byte("abc"), Topic: "zigbee2mqtt/dev", ReceivedAt: time.Now()}
	b.Add(msg)
	if b.Len() != 1 {
		t.Fatalf("expected Len=1, got %d", b.Len())
	}
	all := b.All()
	if string(all[0].Payload) != "abc" {
		t.Errorf("payload mismatch: got %q", all[0].Payload)
	}
	if all[0].Topic != "zigbee2mqtt/dev" {
		t.Errorf("topic mismatch: got %q", all[0].Topic)
	}
}

func TestAdd_Multiple(t *testing.T) {
	b := New()
	msgs := []IRMessage{
		{Payload: []byte("first")},
		{Payload: []byte("second")},
		{Payload: []byte("third")},
	}
	for _, m := range msgs {
		b.Add(m)
	}
	if b.Len() != 3 {
		t.Fatalf("expected Len=3, got %d", b.Len())
	}
	all := b.All()
	for i, m := range msgs {
		if string(all[i].Payload) != string(m.Payload) {
			t.Errorf("index %d: got %q, want %q", i, all[i].Payload, m.Payload)
		}
	}
}

func TestRemove_First(t *testing.T) {
	b := New()
	b.Add(IRMessage{Payload: []byte("a")})
	b.Add(IRMessage{Payload: []byte("b")})
	b.Add(IRMessage{Payload: []byte("c")})
	b.Remove(0)
	if b.Len() != 2 {
		t.Fatalf("expected Len=2, got %d", b.Len())
	}
	all := b.All()
	if string(all[0].Payload) != "b" {
		t.Errorf("expected first to be 'b', got %q", all[0].Payload)
	}
}

func TestRemove_Middle(t *testing.T) {
	b := New()
	b.Add(IRMessage{Payload: []byte("a")})
	b.Add(IRMessage{Payload: []byte("b")})
	b.Add(IRMessage{Payload: []byte("c")})
	b.Remove(1)
	all := b.All()
	if b.Len() != 2 {
		t.Fatalf("expected Len=2, got %d", b.Len())
	}
	if string(all[0].Payload) != "a" || string(all[1].Payload) != "c" {
		t.Errorf("unexpected order: %q %q", all[0].Payload, all[1].Payload)
	}
}

func TestRemove_Last(t *testing.T) {
	b := New()
	b.Add(IRMessage{Payload: []byte("a")})
	b.Add(IRMessage{Payload: []byte("b")})
	b.Remove(1)
	if b.Len() != 1 {
		t.Fatalf("expected Len=1, got %d", b.Len())
	}
	if string(b.All()[0].Payload) != "a" {
		t.Errorf("expected 'a', got %q", b.All()[0].Payload)
	}
}

func TestRemove_NegativeIndex(t *testing.T) {
	b := New()
	b.Add(IRMessage{Payload: []byte("a")})
	b.Remove(-1)
	if b.Len() != 1 {
		t.Fatalf("expected Len=1, got %d", b.Len())
	}
}

func TestRemove_IndexEqualLen(t *testing.T) {
	b := New()
	b.Add(IRMessage{Payload: []byte("a")})
	b.Remove(1) // equal to len, out of range
	if b.Len() != 1 {
		t.Fatalf("expected Len=1, got %d", b.Len())
	}
}

func TestClear(t *testing.T) {
	b := New()
	b.Add(IRMessage{Payload: []byte("a")})
	b.Add(IRMessage{Payload: []byte("b")})
	b.Clear()
	if b.Len() != 0 {
		t.Fatalf("expected Len=0 after Clear, got %d", b.Len())
	}
	if len(b.All()) != 0 {
		t.Fatal("expected empty slice after Clear")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	b := New()
	b.Add(IRMessage{Payload: []byte("original")})
	snapshot := b.All()
	snapshot[0].Payload = []byte("mutated")
	all := b.All()
	if string(all[0].Payload) != "original" {
		t.Errorf("buffer was mutated via returned slice: got %q", all[0].Payload)
	}
}

func TestConcurrentAddAndRead(t *testing.T) {
	b := New()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Add(IRMessage{Payload: []byte("x")})
		}()
	}
	// Read concurrently while goroutines write.
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()
	for {
		select {
		case <-done:
			return
		default:
			_ = b.All()
			_ = b.Len()
		}
	}
}
