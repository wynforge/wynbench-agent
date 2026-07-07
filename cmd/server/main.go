// Command server starts the Wynbench backend engine.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/wynforge/wynbench-agent/api"
	"github.com/wynforge/wynbench-agent/core"
	httpplugin "github.com/wynforge/wynbench-agent/plugins/http"
	sqlplugin "github.com/wynforge/wynbench-agent/plugins/sql"
)

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on")
	flag.Parse()

	// Register built-in protocol plugins.
	core.Register(httpplugin.New())
	core.Register(sqlplugin.New())

	// Initialise in-memory stores and the workflow engine.
	cs := core.NewConnectionStore()
	ws := core.NewWorkflowStore()
	engine := core.NewEngine(cs)

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
