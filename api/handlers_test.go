package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oswryn/wynbench-agent/api"
	"github.com/oswryn/wynbench-agent/core"
)

// stubPlugin is a no-op plugin used across API handler tests.
type stubPlugin struct{}

func (s *stubPlugin) Name() string                           { return "stub" }
func (s *stubPlugin) Configure(_ map[string]any) error       { return nil }
func (s *stubPlugin) Execute(a core.Action) (core.Result, error) {
	return core.Result{Success: true, Data: map[string]any{"echo": a.Params}}, nil
}

func init() {
	core.Register(&stubPlugin{})
}

func newTestServer() (*api.Server, *http.ServeMux) {
	cs := core.NewConnectionStore()
	ws := core.NewWorkflowStore()
	eng := core.NewEngine(cs)
	srv := api.NewServer(cs, ws, eng)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	return srv, mux
}

func postJSON(t *testing.T, mux http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// /connections
// ---------------------------------------------------------------------------

func TestCreateConnection_Created(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/connections", map[string]any{
		"id": "c1", "name": "demo", "protocol": "stub",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateConnection_BadRequest(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/connections", map[string]any{"name": "no-id"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateConnection_Conflict(t *testing.T) {
	_, mux := newTestServer()
	body := map[string]any{"id": "dup", "protocol": "stub"}
	postJSON(t, mux, "/connections", body)
	w := postJSON(t, mux, "/connections", body)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestListConnections(t *testing.T) {
	_, mux := newTestServer()
	postJSON(t, mux, "/connections", map[string]any{"id": "l1", "protocol": "stub"})

	req := httptest.NewRequest(http.MethodGet, "/connections", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []core.Connection
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(list))
	}
}

func TestDeleteConnection(t *testing.T) {
	_, mux := newTestServer()
	postJSON(t, mux, "/connections", map[string]any{"id": "del1", "protocol": "stub"})

	req := httptest.NewRequest(http.MethodDelete, "/connections/del1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteConnection_NotFound(t *testing.T) {
	_, mux := newTestServer()
	req := httptest.NewRequest(http.MethodDelete, "/connections/nope", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// /actions/execute
// ---------------------------------------------------------------------------

func TestExecuteAction_OK(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/actions/execute", map[string]any{
		"plugin": "stub",
		"params": map[string]any{"key": "value"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestExecuteAction_UnknownPlugin(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/actions/execute", map[string]any{
		"plugin": "nonexistent",
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestExecuteAction_MissingPlugin(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/actions/execute", map[string]any{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// /workflows/run
// ---------------------------------------------------------------------------

func TestRunWorkflow_Inline(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/workflows/run", map[string]any{
		"name": "test",
		"steps": []map[string]any{
			{"name": "s1", "action": map[string]any{"plugin": "stub", "params": map[string]any{}}},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var run core.WorkflowRun
	if err := json.NewDecoder(w.Body).Decode(&run); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !run.Success {
		t.Fatal("expected workflow run to succeed")
	}
}

func TestRunWorkflow_MissingSteps(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/workflows/run", map[string]any{"name": "empty"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRunWorkflow_StoredNotFound(t *testing.T) {
	_, mux := newTestServer()
	w := postJSON(t, mux, "/workflows/run", map[string]any{"id": "no-such-id"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
