package httpplugin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oswryn/wynbench-agent/core"
	httpplugin "github.com/oswryn/wynbench-agent/plugins/http"
)

func TestHTTPPlugin_Name(t *testing.T) {
	p := httpplugin.New()
	if p.Name() != "http" {
		t.Fatalf("expected name 'http', got %q", p.Name())
	}
}

func TestHTTPPlugin_GET_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"hello":"world"}`))
	}))
	defer ts.Close()

	p := httpplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "http",
		Params: map[string]any{"url": ts.URL},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got: %s", result.Error)
	}
	if result.Data["status_code"].(int) != http.StatusOK {
		t.Fatalf("unexpected status code: %v", result.Data["status_code"])
	}
	// JSON body should be parsed automatically.
	if result.Data["json"] == nil {
		t.Fatal("expected json key in data")
	}
}

func TestHTTPPlugin_GET_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	p := httpplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "http",
		Params: map[string]any{"url": ts.URL},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure for 404 response")
	}
}

func TestHTTPPlugin_MissingURL(t *testing.T) {
	p := httpplugin.New()
	result, _ := p.Execute(core.Action{Plugin: "http", Params: map[string]any{}})
	if result.Success {
		t.Fatal("expected failure when url param is missing")
	}
}

func TestHTTPPlugin_InvalidScheme(t *testing.T) {
	p := httpplugin.New()
	result, _ := p.Execute(core.Action{
		Plugin: "http",
		Params: map[string]any{"url": "file:///etc/passwd"},
	})
	if result.Success {
		t.Fatal("expected failure for non-http scheme")
	}
}

func TestHTTPPlugin_POST(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	p := httpplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "http",
		Params: map[string]any{"url": ts.URL, "method": "POST", "body": `{"a":1}`},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got: %s", result.Error)
	}
}
