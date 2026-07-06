// Command server starts the Wynbench backend engine.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/oswryn/wynbench-agent/api"
	"github.com/oswryn/wynbench-agent/core"
	httpplugin "github.com/oswryn/wynbench-agent/plugins/http"
	sqlplugin "github.com/oswryn/wynbench-agent/plugins/sql"
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

	log.Printf("wynbench-agent listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
