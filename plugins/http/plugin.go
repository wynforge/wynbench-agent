// Package httpplugin implements the "http" protocol plugin for Wynbench.
//
// It performs basic HTTP GET and POST requests. The plugin is intentionally
// simple; production-grade features (auth, TLS client certs, retries) can be
// added later.
package httpplugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/oswryn/wynbench-agent/core"
)

// Plugin is the HTTP protocol plugin.
type Plugin struct {
	client *http.Client
}

// New returns a ready-to-use HTTP Plugin with default settings.
func New() *Plugin {
	return &Plugin{client: &http.Client{}}
}

// Name returns the protocol identifier used to look up this plugin.
func (p *Plugin) Name() string { return "http" }

// Configure accepts optional connection-level settings. Currently unused but
// required to satisfy the core.Plugin interface.
func (p *Plugin) Configure(_ map[string]any) error { return nil }

// Execute dispatches an HTTP request described by action.Params.
//
// Required params:
//
//	"url"    (string) – the target URL
//
// Optional params:
//
//	"method" (string) – HTTP method; defaults to "GET"
//	"body"   (string) – request body for POST/PUT
func (p *Plugin) Execute(action core.Action) (core.Result, error) {
	urlVal, ok := action.Params["url"]
	if !ok {
		return core.Result{Success: false, Error: "missing param: url"}, nil
	}
	url, ok := urlVal.(string)
	if !ok || url == "" {
		return core.Result{Success: false, Error: "param url must be a non-empty string"}, nil
	}

	method := "GET"
	if m, ok := action.Params["method"].(string); ok && m != "" {
		method = strings.ToUpper(m)
	}

	var bodyReader io.Reader
	if b, ok := action.Params["body"].(string); ok && b != "" {
		bodyReader = strings.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return core.Result{Success: false, Error: fmt.Sprintf("failed to build request: %v", err)}, nil
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return core.Result{Success: false, Error: fmt.Sprintf("request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)

	data := map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(rawBody),
	}

	// Attempt to parse JSON body for richer downstream use.
	var parsed any
	if err := json.Unmarshal(rawBody, &parsed); err == nil {
		data["json"] = parsed
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	result := core.Result{Success: success, Data: data}
	if !success {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return result, nil
}
