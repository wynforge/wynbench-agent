package core_test

import (
	"testing"

	"github.com/wynforge/wynbench-agent/core"
)

// enginePlugin is a deterministic Plugin used in engine tests.
type enginePlugin struct{}

func (e *enginePlugin) Name() string                     { return "engine-test" }
func (e *enginePlugin) Configure(_ map[string]any) error { return nil }
func (e *enginePlugin) Execute(a core.Action) (core.Result, error) {
	return core.Result{Success: true, Data: map[string]any{"params": a.Params}}, nil
}

func init() {
	core.Register(&enginePlugin{})
}

func TestExecuteAction_OK(t *testing.T) {
	cs := core.NewConnectionStore()
	eng := core.NewEngine(cs)

	result, err := eng.ExecuteAction(core.Action{
		Plugin: "engine-test",
		Params: map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
}

func TestExecuteAction_UnknownPlugin(t *testing.T) {
	cs := core.NewConnectionStore()
	eng := core.NewEngine(cs)

	result, err := eng.ExecuteAction(core.Action{Plugin: "no-such-plugin"})
	if err == nil {
		t.Fatal("expected error for unknown plugin")
	}
	if result.Success {
		t.Fatal("expected result.Success to be false")
	}
}

func TestExecuteAction_ConnectionMerge(t *testing.T) {
	cs := core.NewConnectionStore()
	_ = cs.Add(core.Connection{
		ID:       "conn1",
		Protocol: "engine-test",
		Config:   map[string]any{"base": "from-conn"},
	})

	eng := core.NewEngine(cs)
	result, err := eng.ExecuteAction(core.Action{
		Plugin:       "engine-test",
		ConnectionID: "conn1",
		Params:       map[string]any{"override": "from-action"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := result.Data["params"].(map[string]any)
	if params["base"] != "from-conn" {
		t.Fatalf("expected 'base' to be 'from-conn', got %v", params["base"])
	}
	if params["override"] != "from-action" {
		t.Fatalf("expected 'override' to be 'from-action', got %v", params["override"])
	}
}

func TestExecuteAction_MissingConnection(t *testing.T) {
	cs := core.NewConnectionStore()
	eng := core.NewEngine(cs)

	result, _ := eng.ExecuteAction(core.Action{
		Plugin:       "engine-test",
		ConnectionID: "missing",
	})
	if result.Success {
		t.Fatal("expected failure for missing connection")
	}
}

func TestRunWorkflow(t *testing.T) {
	cs := core.NewConnectionStore()
	eng := core.NewEngine(cs)

	wf := core.Workflow{
		ID:   "wf-test",
		Name: "test",
		Steps: []core.WorkflowStep{
			{Name: "step1", Action: core.Action{Plugin: "engine-test", Params: map[string]any{"n": 1}}},
			{Name: "step2", Action: core.Action{Plugin: "engine-test", Params: map[string]any{"n": 2}}},
		},
	}

	run := eng.RunWorkflow(wf)
	if !run.Success {
		t.Fatalf("expected workflow run to succeed")
	}
	if len(run.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(run.Results))
	}
}

func TestRunWorkflow_StopsOnFailure(t *testing.T) {
	cs := core.NewConnectionStore()
	eng := core.NewEngine(cs)

	wf := core.Workflow{
		ID: "wf-fail",
		Steps: []core.WorkflowStep{
			{Name: "bad", Action: core.Action{Plugin: "no-such-plugin"}},
			{Name: "ok", Action: core.Action{Plugin: "engine-test"}},
		},
	}

	run := eng.RunWorkflow(wf)
	if run.Success {
		t.Fatal("expected workflow run to fail")
	}
	// Should stop after the first failing step.
	if len(run.Results) != 1 {
		t.Fatalf("expected 1 result (stopped after failure), got %d", len(run.Results))
	}
}
