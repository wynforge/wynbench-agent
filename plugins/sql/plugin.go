// Package sqlplugin implements the "sql" protocol plugin for Wynbench.
//
// This is a placeholder implementation. It validates the expected parameters
// and returns a stub response. Actual database connectivity (e.g. via
// database/sql + a driver) can be wired in later without changing the plugin
// interface.
package sqlplugin

import (
	"fmt"

	"github.com/wynforge/wynbench-agent/core"
)

// Plugin is the SQL protocol plugin.
type Plugin struct{}

// New returns a ready-to-use SQL Plugin.
func New() *Plugin { return &Plugin{} }

// Name returns the protocol identifier used to look up this plugin.
func (p *Plugin) Name() string { return "sql" }

// Configure accepts optional connection-level settings such as a connection
// string. The placeholder implementation performs basic validation only.
func (p *Plugin) Configure(cfg map[string]any) error {
	if dsn, ok := connectionString(cfg); ok && dsn != "" {
		return nil
	}
	// A missing connection string is not fatal at configure time; it will
	// surface at Execute.
	return nil
}

// connectionString reads the SQL connection string from cfg, accepting
// either the current "connectionString" key or the legacy "dsn" key.
func connectionString(cfg map[string]any) (string, bool) {
	if v, ok := cfg["connectionString"].(string); ok && v != "" {
		return v, true
	}
	if v, ok := cfg["dsn"].(string); ok && v != "" {
		return v, true
	}
	return "", false
}

// Execute validates SQL action parameters and returns a stub result.
//
// Expected params:
//
//	"query" (string) – the SQL query to execute
//
// Optional params:
//
//	"connectionString" (string) – connection string (overrides connection config)
//	"dsn"               (string) – legacy alias for "connectionString"
func (p *Plugin) Execute(action core.Action) (core.Result, error) {
	query, ok := action.Params["query"].(string)
	if !ok || query == "" {
		return core.Result{Success: false, Error: "missing param: query"}, nil
	}

	// Placeholder: log the intent and return a stub acknowledgement.
	return core.Result{
		Success: true,
		Data: map[string]any{
			"message": fmt.Sprintf("SQL plugin (stub): query received: %q", query),
			"rows":    []map[string]any{},
		},
	}, nil
}
