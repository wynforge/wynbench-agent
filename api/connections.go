package api

import (
	"encoding/json"
	"net/http"

	"github.com/wynforge/wynbench-agent/core"
)

// createConnection handles POST /connections.
// Body: { "id": "...", "name": "...", "protocol": "...", "config": {...} }
func (s *Server) createConnection(w http.ResponseWriter, r *http.Request) {
	var c core.Connection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if c.ID == "" || c.Protocol == "" {
		writeError(w, http.StatusBadRequest, "id and protocol are required")
		return
	}
	if err := s.connections.Add(c); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// listConnections handles GET /connections.
func (s *Server) listConnections(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.connections.List())
}

// deleteConnection handles DELETE /connections/{id}.
func (s *Server) deleteConnection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing connection id")
		return
	}
	if err := s.connections.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
