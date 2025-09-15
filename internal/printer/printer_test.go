package printer

import (
	"context"
	"testing"
	"time"
)

func QueueTest(t *testing.T) {
	p := New(context.Background())
	if p == nil {
		t.Fatal("Printer is nil")
	}
	p.Enabled = false
	err := p.PrintTask(1, "Test Task", "This is a test task", 0, "Tester", time.Now())
	if err == nil {
		t.Fatal("Expected error when printing with disabled printer")
	}
	if len(p.queue) != 1 {
		t.Fatalf("Expected queue length 1, got %d", len(p.queue))
	}
	err = p.PrintTask(2, "Test Task", "This is a test task", 0, "Tester", time.Now())
	if err == nil {
		t.Fatal("Expected error when printing with disabled printer")
	}
	if len(p.queue) != 2 {
		t.Fatalf("Expected queue length 1, got %d", len(p.queue))
	}
	p.queue = nil
}
