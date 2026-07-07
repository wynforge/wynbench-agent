package api

import (
	"net/http"
	"sort"

	"github.com/wynforge/wynbench-agent/core"
)

// health handles GET /health.
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	plugins := core.Registered()
	sort.Strings(plugins)

	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"plugins":     plugins,
		"connections": len(s.connections.List()),
	})
}
