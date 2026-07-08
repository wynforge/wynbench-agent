package core_test

import (
	"testing"

	"github.com/wynforge/wynbench-agent/core"
)

// fakePlugin is a minimal Plugin used in registry tests.
type fakePlugin struct{ name string }

func (f *fakePlugin) Name() string                     { return f.name }
func (f *fakePlugin) Configure(_ map[string]any) error { return nil }
func (f *fakePlugin) Execute(a core.Action) (core.Result, error) {
	return core.Result{Success: true, Data: map[string]any{"echo": a.Params}}, nil
}

func TestRegisterAndGet(t *testing.T) {
	p := &fakePlugin{name: "fake"}
	core.Register(p)

	got, ok := core.Get("fake")
	if !ok {
		t.Fatal("expected to find registered plugin 'fake'")
	}
	if got.Name() != "fake" {
		t.Fatalf("expected name 'fake', got %q", got.Name())
	}
}

func TestGetUnknown(t *testing.T) {
	_, ok := core.Get("does-not-exist")
	if ok {
		t.Fatal("expected false for unknown plugin")
	}
}

func TestRegistered(t *testing.T) {
	names := core.Registered()
	found := false
	for _, n := range names {
		if n == "fake" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected 'fake' in Registered()")
	}
}

func TestRegisterPanicOnDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	// 'fake' was registered above; registering again must panic.
	core.Register(&fakePlugin{name: "fake"})
}
