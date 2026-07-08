package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/wynforge/wynbench-agent/core"
)

// createWorkflow handles POST /workflows.
// Body: { "id": "...", "name": "...", "steps": [...] }
func (s *Server) createWorkflow(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.ID == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "id and name are required")
		return
	}
	if len(req.Steps) == 0 {
		writeError(w, http.StatusBadRequest, "at least one step is required")
		return
	}

	wf := core.Workflow{
		ID:        req.ID,
		Name:      req.Name,
		Steps:     req.Steps,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.workflows.Add(wf); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	s.persist()
	writeJSON(w, http.StatusCreated, wf)
}

// listWorkflows handles GET /workflows.
func (s *Server) listWorkflows(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.workflows.List())
}

// updateWorkflow handles PUT /workflows/{id}.
func (s *Server) updateWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing workflow id")
		return
	}

	var req UpdateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Steps) == 0 {
		writeError(w, http.StatusBadRequest, "at least one step is required")
		return
	}

	stored, ok := s.workflows.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "workflow not found")
		return
	}

	updated := core.Workflow{
		ID:        stored.ID,
		Name:      req.Name,
		Steps:     req.Steps,
		CreatedAt: stored.CreatedAt,
	}
	if err := s.workflows.Update(updated); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.persist()
	writeJSON(w, http.StatusOK, updated)
}

// deleteWorkflow handles DELETE /workflows/{id}.
func (s *Server) deleteWorkflow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing workflow id")
		return
	}
	if err := s.workflows.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.persist()
	w.WriteHeader(http.StatusNoContent)
}

// runWorkflow handles POST /workflows/run.
//
// The client may either supply a full inline Workflow definition or a
// previously stored workflow ID.
//
//	{ "id": "<stored-id>" }                 – run a stored workflow
//	{ "name": "...", "steps": [...] }       – run an inline (ad-hoc) workflow
func (s *Server) runWorkflow(w http.ResponseWriter, r *http.Request) {
	var req RunWorkflowRequest
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
