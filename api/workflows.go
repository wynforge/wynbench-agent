package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/oswryn/wynbench-agent/core"
)

// runWorkflow handles POST /workflows/run.
//
// The client may either supply a full inline Workflow definition or a
// previously stored workflow ID.
//
//	{ "id": "<stored-id>" }                 – run a stored workflow
//	{ "name": "...", "steps": [...] }       – run an inline (ad-hoc) workflow
func (s *Server) runWorkflow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		// Reference to a stored workflow.
		ID string `json:"id"`
		// Inline workflow fields.
		Name  string             `json:"name"`
		Steps []core.WorkflowStep `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var wf core.Workflow

	if req.ID != "" {
		stored, ok := s.workflows.Get(req.ID)
		if !ok {
			writeError(w, http.StatusNotFound, "workflow not found")
			return
		}
		wf = stored
	} else {
		if len(req.Steps) == 0 {
			writeError(w, http.StatusBadRequest, "steps are required for an inline workflow")
			return
		}
		wf = core.Workflow{
			ID:        "inline",
			Name:      req.Name,
			Steps:     req.Steps,
			CreatedAt: time.Now().UTC(),
		}
	}

	run := s.engine.RunWorkflow(wf)
	writeJSON(w, http.StatusOK, run)
}
