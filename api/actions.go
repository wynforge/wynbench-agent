package api

import (
	"encoding/json"
	"net/http"

	"github.com/oswryn/wynbench-agent/core"
)

// executeAction handles POST /actions/execute.
// Body: a core.Action JSON object.
func (s *Server) executeAction(w http.ResponseWriter, r *http.Request) {
	var a core.Action
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if a.Plugin == "" {
		writeError(w, http.StatusBadRequest, "plugin is required")
		return
	}

	result, err := s.engine.ExecuteAction(a)
	if err != nil {
		// Engine errors (e.g. unknown plugin) surface as 400/422.
		writeJSON(w, http.StatusUnprocessableEntity, result)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
