// Command server starts the Wynbench backend engine.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/wynforge/wynbench-agent/api"
	"github.com/wynforge/wynbench-agent/config"
	"github.com/wynforge/wynbench-agent/core"
	httpplugin "github.com/wynforge/wynbench-agent/plugins/http"
	kafkaplugin "github.com/wynforge/wynbench-agent/plugins/kafka"
	sqlplugin "github.com/wynforge/wynbench-agent/plugins/sql"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on")
	flag.Parse()

	// Register built-in protocol plugins.
	core.Register(httpplugin.New())
	core.Register(sqlplugin.New())
	core.Register(kafkaplugin.New())

	// Initialise in-memory stores and the workflow engine.
	cs := core.NewConnectionStore()
	ws := core.NewWorkflowStore()
	engine := core.NewEngine(cs)

	// Restore any previously saved connections and workflows from disk.
	if snap, err := config.Load(); err != nil {
		log.Printf("warning: failed to load persisted config: %v", err)
	} else {
		for _, c := range snap.Connections {
			if err := cs.Add(c); err != nil {
				log.Printf("warning: failed to restore connection %q: %v", c.ID, err)
			}
		}
		for _, wf := range snap.Workflows {
			if err := ws.Add(wf); err != nil {
				log.Printf("warning: failed to restore workflow %q: %v", wf.ID, err)
			}
		}
		if path, err := config.Path(); err == nil {
			log.Printf("config store: %s", path)
		}
	}

	// Wire HTTP routes.
	mux := http.NewServeMux()
	srv := api.NewServer(cs, ws, engine)
	srv.RegisterRoutes(mux)
	handler := withCORS(mux)

	log.Printf("wynbench-agent listening on %s", *addr)
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Accept")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
