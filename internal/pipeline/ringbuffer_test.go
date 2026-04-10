package pipeline

import "testing"

func TestRingBufferBasic(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Write("line 1")
	rb.Write("line 2")
	rb.Write("line 3")

	lines := rb.Lines()
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "line 1" || lines[2] != "line 3" {
		t.Errorf("wrong order: %v", lines)
	}
}

func TestRingBufferOverwrite(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Write("a")
	rb.Write("b")
	rb.Write("c")
	rb.Write("d") // should overwrite "a"

	lines := rb.Lines()
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "b" || lines[1] != "c" || lines[2] != "d" {
		t.Errorf("expected [b c d], got %v", lines)
	}
}

func TestRingBufferEmpty(t *testing.T) {
	rb := NewRingBuffer(5)
	lines := rb.Lines()
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0", len(lines))
	}
}

func TestRingBufferPartialFill(t *testing.T) {
	rb := NewRingBuffer(10)
	rb.Write("only one")
	lines := rb.Lines()
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if lines[0] != "only one" {
		t.Errorf("got %q, want %q", lines[0], "only one")
	}
}
