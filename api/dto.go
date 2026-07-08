package api

import "github.com/wynforge/wynbench-agent/core"

// CreateWorkflowRequest is the request body for POST /workflows.
type CreateWorkflowRequest struct {
	ID    string              `json:"id"`
	Name  string              `json:"name"`
	Steps []core.WorkflowStep `json:"steps"`
}

// UpdateWorkflowRequest is the request body for PUT /workflows/{id}.
type UpdateWorkflowRequest struct {
	Name  string              `json:"name"`
	Steps []core.WorkflowStep `json:"steps"`
}

// RunWorkflowRequest is the request body for POST /workflows/run.
//
// The client may either supply a full inline Workflow definition or a
// previously stored workflow ID.
//
//	{ "id": "<stored-id>" }                 – run a stored workflow
//	{ "name": "...", "steps": [...] }       – run an inline (ad-hoc) workflow
type RunWorkflowRequest struct {
	// ID references a stored workflow. When set, Name/Steps are ignored.
	ID string `json:"id"`
	// Name and Steps describe an inline (ad-hoc) workflow run.
	Name  string              `json:"name"`
	Steps []core.WorkflowStep `json:"steps"`
}
