// Package core defines the fundamental types used throughout the Wynbench
// backend engine: Connection, Action, Result, and Workflow.
package core

import "time"

// Connection represents a named, protocol-specific connection configuration
// stored in the in-memory registry.
type Connection struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Protocol string         `json:"protocol"`
	Config   map[string]any `json:"config"`
}

// Action describes a single operation to be executed against a protocol plugin.
type Action struct {
	// Plugin is the protocol name that should handle this action (e.g. "http", "sql").
	Plugin string `json:"plugin"`
	// ConnectionID is the optional ID of a stored Connection whose Config is
	// merged into Params before execution.
	ConnectionID string `json:"connection_id,omitempty"`
	// Params holds action-specific parameters passed directly to the plugin.
	Params map[string]any `json:"params"`
}

// Result holds the outcome of executing an Action.
type Result struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// WorkflowStep pairs an Action with a human-readable name.
type WorkflowStep struct {
	Name   string `json:"name"`
	Action Action `json:"action"`
}

// Workflow is an ordered sequence of steps that the engine executes in series.
type Workflow struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Steps     []WorkflowStep `json:"steps"`
	CreatedAt time.Time      `json:"created_at"`
}

// WorkflowRun captures the outcome of running a Workflow.
type WorkflowRun struct {
	WorkflowID string   `json:"workflow_id"`
	Results    []Result `json:"results"`
	Success    bool     `json:"success"`
}
