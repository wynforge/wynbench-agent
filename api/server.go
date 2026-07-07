// Package api wires together the HTTP routes for the Wynbench backend engine.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/wynforge/wynbench-agent/core"
)

// Server holds the shared dependencies for all HTTP handlers.
type Server struct {
	connections *core.ConnectionStore
	workflows   *core.WorkflowStore
	engine      *core.Engine
}

// NewServer creates a Server backed by the supplied stores and engine.
func NewServer(cs *core.ConnectionStore, ws *core.WorkflowStore, e *core.Engine) *Server {
	return &Server{connections: cs, workflows: ws, engine: e}
}

// RegisterRoutes attaches all API routes to mux. Using a plain http.ServeMux
// keeps the dependency footprint minimal.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.health)

	mux.HandleFunc("POST /connections", s.createConnection)
	mux.HandleFunc("GET /connections", s.listConnections)
	mux.HandleFunc("DELETE /connections/{id}", s.deleteConnection)

	mux.HandleFunc("POST /actions/execute", s.executeAction)

	mux.HandleFunc("POST /workflows/run", s.runWorkflow)
}

// writeJSON serialises v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError sends a JSON error envelope.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
