package kafkaplugin_test

import (
	"testing"

	"github.com/wynforge/wynbench-agent/core"
	kafkaplugin "github.com/wynforge/wynbench-agent/plugins/kafka"
)

func TestKafkaPlugin_Name(t *testing.T) {
	p := kafkaplugin.New()
	if p.Name() != "kafka" {
		t.Fatalf("expected name 'kafka', got %q", p.Name())
	}
}

func TestKafkaPlugin_Configure(t *testing.T) {
	p := kafkaplugin.New()
	if err := p.Configure(map[string]any{"brokers": "localhost:9092"}); err != nil {
		t.Fatalf("unexpected configure error: %v", err)
	}
}

func TestKafkaPlugin_MissingBrokers(t *testing.T) {
	p := kafkaplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "kafka",
		Params: map[string]any{"topic": "test", "value": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when brokers param is missing")
	}
}

func TestKafkaPlugin_MissingTopic(t *testing.T) {
	p := kafkaplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "kafka",
		Params: map[string]any{"brokers": "localhost:9092", "value": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when topic param is missing")
	}
}

func TestKafkaPlugin_MissingValue(t *testing.T) {
	p := kafkaplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "kafka",
		Params: map[string]any{"brokers": "localhost:9092", "topic": "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when value param is missing")
	}
}

func TestKafkaPlugin_BlankBrokers(t *testing.T) {
	p := kafkaplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "kafka",
		Params: map[string]any{"brokers": "  , ,", "topic": "test", "value": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when brokers list has no usable addresses")
	}
}

func TestKafkaPlugin_ProduceUnreachableBroker(t *testing.T) {
	// No live broker is available in unit tests; verify that a connection
	// failure is surfaced as a failed Result rather than a panic or hang.
	p := kafkaplugin.New()
	result, err := p.Execute(core.Action{
		Plugin: "kafka",
		Params: map[string]any{
			"brokers": "127.0.0.1:1", // reserved/unused port, connection refused fast
			"topic":   "test",
			"value":   "hello",
			"key":     "k1",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when broker is unreachable")
	}
	if result.Error == "" {
		t.Fatal("expected an error message describing the produce failure")
	}
}
