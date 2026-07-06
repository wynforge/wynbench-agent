package core_test

import (
	"testing"

	"github.com/oswryn/wynbench-agent/core"
)

func TestConnectionStore(t *testing.T) {
	s := core.NewConnectionStore()

	c := core.Connection{ID: "c1", Name: "test", Protocol: "http"}

	if err := s.Add(c); err != nil {
		t.Fatalf("Add: unexpected error: %v", err)
	}

	// Duplicate add must fail.
	if err := s.Add(c); err == nil {
		t.Fatal("Add duplicate: expected error, got nil")
	}

	got, ok := s.Get("c1")
	if !ok {
		t.Fatal("Get: expected to find c1")
	}
	if got.ID != "c1" {
		t.Fatalf("Get: expected id 'c1', got %q", got.ID)
	}

	list := s.List()
	if len(list) != 1 {
		t.Fatalf("List: expected 1 item, got %d", len(list))
	}

	if err := s.Delete("c1"); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}
	if _, ok := s.Get("c1"); ok {
		t.Fatal("Get after Delete: expected not found")
	}

	// Delete non-existent must fail.
	if err := s.Delete("c1"); err == nil {
		t.Fatal("Delete non-existent: expected error, got nil")
	}
}

func TestWorkflowStore(t *testing.T) {
	s := core.NewWorkflowStore()

	wf := core.Workflow{ID: "wf1", Name: "smoke"}

	if err := s.Add(wf); err != nil {
		t.Fatalf("Add: unexpected error: %v", err)
	}

	// Duplicate add must fail.
	if err := s.Add(wf); err == nil {
		t.Fatal("Add duplicate: expected error, got nil")
	}

	got, ok := s.Get("wf1")
	if !ok {
		t.Fatal("Get: expected to find wf1")
	}
	if got.ID != "wf1" {
		t.Fatalf("Get: expected id 'wf1', got %q", got.ID)
	}

	list := s.List()
	if len(list) != 1 {
		t.Fatalf("List: expected 1 item, got %d", len(list))
	}

	if err := s.Delete("wf1"); err != nil {
		t.Fatalf("Delete: unexpected error: %v", err)
	}
}
