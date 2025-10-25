package ui

import (
	"testing"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

func TestNewResultBuffer(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		expected int
	}{
		{"valid capacity", 100, 100},
		{"zero capacity uses default", 0, DefaultResultBufferSize},
		{"negative capacity uses default", -1, DefaultResultBufferSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewResultBuffer(tt.capacity)
			if rb.capacity != tt.expected {
				t.Errorf("expected capacity %d, got %d", tt.expected, rb.capacity)
			}
			if rb.Len() != 0 {
				t.Errorf("expected length 0, got %d", rb.Len())
			}
		})
	}
}

func TestResultBuffer_AppendAndItems(t *testing.T) {
	rb := NewResultBuffer(3)

	// Add first result
	r1 := core.ResultEvent{Host: "host1", Port: 80, State: core.StateOpen}
	rb.Append(r1)

	if rb.Len() != 1 {
		t.Errorf("expected length 1, got %d", rb.Len())
	}

	items := rb.Items()
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if items[0].Host != "host1" {
		t.Errorf("expected host1, got %s", items[0].Host)
	}

	// Add two more results
	r2 := core.ResultEvent{Host: "host2", Port: 443, State: core.StateClosed}
	r3 := core.ResultEvent{Host: "host3", Port: 22, State: core.StateFiltered}
	rb.Append(r2)
	rb.Append(r3)

	if rb.Len() != 3 {
		t.Errorf("expected length 3, got %d", rb.Len())
	}

	items = rb.Items()
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	// Verify order (oldest to newest)
	if items[0].Host != "host1" || items[1].Host != "host2" || items[2].Host != "host3" {
		t.Error("items not in correct order")
	}
}

func TestResultBuffer_WrapAround(t *testing.T) {
	rb := NewResultBuffer(3)

	// Fill buffer
	r1 := core.ResultEvent{Host: "host1", Port: 80}
	r2 := core.ResultEvent{Host: "host2", Port: 443}
	r3 := core.ResultEvent{Host: "host3", Port: 22}
	rb.Append(r1)
	rb.Append(r2)
	rb.Append(r3)

	// Add one more - should evict host1
	r4 := core.ResultEvent{Host: "host4", Port: 3306}
	rb.Append(r4)

	if rb.Len() != 3 {
		t.Errorf("expected length 3 after wrap, got %d", rb.Len())
	}

	items := rb.Items()
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	// Should have host2, host3, host4 (host1 evicted)
	hosts := []string{items[0].Host, items[1].Host, items[2].Host}
	expected := []string{"host2", "host3", "host4"}

	for i, host := range hosts {
		if host != expected[i] {
			t.Errorf("at position %d: expected %s, got %s", i, expected[i], host)
		}
	}
}

func TestResultBuffer_EmptyBuffer(t *testing.T) {
	rb := NewResultBuffer(5)

	items := rb.Items()
	if items != nil {
		t.Errorf("expected nil items for empty buffer, got %d items", len(items))
	}

	if rb.Len() != 0 {
		t.Errorf("expected length 0, got %d", rb.Len())
	}
}

func TestResultStats_Add(t *testing.T) {
	stats := NewResultStats()

	// Add results with different states
	r1 := core.ResultEvent{Host: "host1", Port: 80, State: core.StateOpen}
	r2 := core.ResultEvent{Host: "host2", Port: 443, State: core.StateClosed}
	r3 := core.ResultEvent{Host: "host3", Port: 22, State: core.StateFiltered}
	r4 := core.ResultEvent{Host: "host4", Port: 3306, State: core.StateOpen}

	stats.Add(r1)
	stats.Add(r2)
	stats.Add(r3)
	stats.Add(r4)

	total, open, closed, filtered := stats.Totals()

	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}
	if open != 2 {
		t.Errorf("expected open 2, got %d", open)
	}
	if closed != 1 {
		t.Errorf("expected closed 1, got %d", closed)
	}
	if filtered != 1 {
		t.Errorf("expected filtered 1, got %d", filtered)
	}
}

func TestResultStats_Empty(t *testing.T) {
	stats := NewResultStats()

	total, open, closed, filtered := stats.Totals()

	if total != 0 || open != 0 || closed != 0 || filtered != 0 {
		t.Errorf("expected all zeros, got total=%d open=%d closed=%d filtered=%d",
			total, open, closed, filtered)
	}
}

func TestResultBuffer_LargeDataset(t *testing.T) {
	// Test with a larger buffer
	bufferSize := 1000
	rb := NewResultBuffer(bufferSize)

	// Add more results than buffer capacity
	totalResults := 1500
	for i := 0; i < totalResults; i++ {
		rb.Append(core.ResultEvent{
			Host:  "host",
			Port:  uint16(i % 65535),
			State: core.StateOpen,
		})
	}

	// Buffer should only contain last 1000 results
	if rb.Len() != bufferSize {
		t.Errorf("expected length %d, got %d", bufferSize, rb.Len())
	}

	items := rb.Items()
	if len(items) != bufferSize {
		t.Errorf("expected %d items, got %d", bufferSize, len(items))
	}

	// Verify oldest result is the 501st one (index 500)
	if items[0].Port != uint16(500%65535) {
		t.Errorf("expected first port to be %d, got %d", 500%65535, items[0].Port)
	}

	// Verify newest result is the 1500th one (index 1499)
	if items[bufferSize-1].Port != uint16(1499%65535) {
		t.Errorf("expected last port to be %d, got %d", 1499%65535, items[bufferSize-1].Port)
	}
}

func TestResultBuffer_WithStats(t *testing.T) {
	rb := NewResultBuffer(2) // Small buffer
	stats := NewResultStats()

	// Add results exceeding buffer capacity
	r1 := core.ResultEvent{Host: "h1", Port: 80, State: core.StateOpen, Duration: 10 * time.Millisecond}
	r2 := core.ResultEvent{Host: "h2", Port: 443, State: core.StateClosed, Duration: 20 * time.Millisecond}
	r3 := core.ResultEvent{Host: "h3", Port: 22, State: core.StateFiltered, Duration: 15 * time.Millisecond}

	rb.Append(r1)
	stats.Add(r1)

	rb.Append(r2)
	stats.Add(r2)

	rb.Append(r3)
	stats.Add(r3) // r1 is evicted from buffer

	// Buffer should only have 2 items
	if rb.Len() != 2 {
		t.Errorf("expected buffer length 2, got %d", rb.Len())
	}

	// But stats should track all 3
	total, open, closed, filtered := stats.Totals()
	if total != 3 {
		t.Errorf("expected total stats 3, got %d", total)
	}
	if open != 1 || closed != 1 || filtered != 1 {
		t.Errorf("expected 1 of each state, got open=%d closed=%d filtered=%d",
			open, closed, filtered)
	}

	// Buffer should only contain r2 and r3
	items := rb.Items()
	if len(items) != 2 {
		t.Errorf("expected 2 items in buffer, got %d", len(items))
	}
	if items[0].Host != "h2" || items[1].Host != "h3" {
		t.Errorf("buffer should contain h2 and h3, got %s and %s", items[0].Host, items[1].Host)
	}
}
