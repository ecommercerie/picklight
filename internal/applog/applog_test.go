package applog

import "testing"

func TestAddAndGet(t *testing.T) {
	l := New(100)
	l.Info("msg %d", 1)
	l.Warn("warn")
	l.Error("err")
	entries := l.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3, got %d", len(entries))
	}
}

func TestRingBuffer(t *testing.T) {
	l := New(3)
	l.Info("a")
	l.Info("b")
	l.Info("c")
	l.Info("d")
	entries := l.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3, got %d", len(entries))
	}
	if entries[0].Message != "b" {
		t.Errorf("expected 'b', got %s", entries[0].Message)
	}
}

func TestClear(t *testing.T) {
	l := New(100)
	l.Info("test")
	l.Clear()
	if len(l.GetEntries()) != 0 {
		t.Error("expected empty")
	}
}
