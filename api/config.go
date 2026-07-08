package api

import (
	"encoding/json"
	"net/http"

	"github.com/wynforge/wynbench-agent/config"
)

// exportConfig handles GET /config/export.
//
// Returns the full configuration bundle (stored connections and workflows)
// as a downloadable JSON file, sourced from the in-memory stores so the
// export always reflects the latest state even if a save is still pending.
func (s *Server) exportConfig(w http.ResponseWriter, _ *http.Request) {
	snap := config.Snapshot{
		Connections: s.connections.List(),
		Workflows:   s.workflows.List(),
	}

	w.Header().Set("Content-Disposition", `attachment; filename="wynbench-config.json"`)
	writeJSON(w, http.StatusOK, snap)
}

// importConfig handles POST /config/import.
//
// Body: { "connections": [...], "workflows": [...] } – a Snapshot previously
// produced by GET /config/export. Importing replaces the entire connection
// and workflow sets and persists them to disk.
func (s *Server) importConfig(w http.ResponseWriter, r *http.Request) {
	var snap config.Snapshot
	if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	s.connections.ReplaceAll(snap.Connections)
	s.workflows.ReplaceAll(snap.Workflows)
	s.persist()

	writeJSON(w, http.StatusOK, config.Snapshot{
		Connections: s.connections.List(),
		Workflows:   s.workflows.List(),
	})
}

// configPath handles GET /config/path.
//
// Returns the on-disk location of the persisted config file, so the UI can
// display where Wynbench stores its configuration.
func (s *Server) configPath(w http.ResponseWriter, _ *http.Request) {
	path, err := config.Path()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": path})
}
